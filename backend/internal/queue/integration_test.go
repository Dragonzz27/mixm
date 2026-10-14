package queue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
	"mixm-edu/backend/internal/testutil"
)

func TestRabbitMQOutboxAndDuplicateDelivery(t *testing.T) {
	if os.Getenv("TEST_RABBITMQ_URL") == "" || os.Getenv("TEST_REDIS_URL") == "" {
		t.Skip("set TEST_RABBITMQ_URL and TEST_REDIS_URL for queue integration")
	}
	db := testutil.Database(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	opts, e := redis.ParseURL(os.Getenv("TEST_REDIS_URL"))
	if e != nil {
		t.Fatal(e)
	}
	cache := redis.NewClient(opts)
	defer cache.Close()
	name := "test_" + uuid.NewString()
	cfg := config.Config{RabbitURL: os.Getenv("TEST_RABBITMQ_URL"), Queue: name, PVCCapacity: "10Gi", PVCAccessMode: "ReadWriteMany", StorageClass: "longhorn"}
	conn, ch, e := channel(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()
	defer ch.Close()
	defer ch.QueueDelete(name, false, false, false)
	u, e := store.Write(ctx, db, "users", "", store.Row{"username": "queue-test", "hashed_password": "unused"})
	if e != nil {
		t.Fatal(e)
	}
	v := store.Row{"name": "test", "user_id": u["id"], "image": "ubuntu:24.04", "config": `{"password":"secret"}`}
	if e = cluster.Validate(v); e != nil {
		t.Fatal(e)
	}
	inst, e := store.Write(ctx, db, "instances", "", v)
	if e != nil {
		t.Fatal(e)
	}
	task, e := store.Write(ctx, db, "instance_tasks", "", store.Row{"user_id": u["id"], "instance_id": inst["id"], "payload": v})
	if e != nil {
		t.Fatal(e)
	}
	id := store.String(task, "id")
	defer cache.Del(context.Background(), "instance_task:"+id)
	k := kubefake.NewSimpleClientset()
	k.PrependReactor("get", "deployments", func(a ktesting.Action) (bool, runtime.Object, error) {
		obj, e := k.Tracker().Get(appsv1.SchemeGroupVersion.WithResource("deployments"), a.GetNamespace(), a.(ktesting.GetAction).GetName())
		if e != nil {
			return true, nil, e
		}
		d := obj.(*appsv1.Deployment).DeepCopy()
		d.Status.ObservedGeneration = d.Generation
		d.Status.AvailableReplicas = 1
		d.Status.ReadyReplicas = 1
		return true, d, nil
	})
	client := &cluster.Client{K: k, Dynamic: fake.NewSimpleDynamicClient(runtime.NewScheme()), Config: cfg}
	publisherDone, workerDone := make(chan struct{}), make(chan error, 1)
	go func() { Dispatch(ctx, cfg, db); close(publisherDone) }()
	go func() { workerDone <- Work(ctx, cfg, db, cache, client) }()
	t.Cleanup(func() {
		cancel()
		<-publisherDone
		select {
		case <-workerDone:
		case <-time.After(10 * time.Second):
			t.Error("worker shutdown timeout")
		}
	})
	deadline := time.Now().Add(15 * time.Second)
	for {
		row, e := store.Get(ctx, db, "instance_tasks", id)
		if e != nil {
			t.Fatal(e)
		}
		if row["status"] == "completed" {
			break
		}
		if row["status"] == "failed" {
			t.Fatalf("task failed: %v", row)
		}
		if time.Now().After(deadline) {
			t.Fatalf("task not completed: %v", row)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if e = ch.PublishWithContext(ctx, "", name, false, false, amqp.Publishing{Body: []byte(id), DeliveryMode: amqp.Persistent}); e != nil {
		t.Fatal(e)
	}
	if e = process(ctx, db, cache, client, id); e != nil {
		t.Fatal(e)
	}
	row, e := store.Get(ctx, db, "instance_tasks", id)
	if e != nil || store.Number(row, "attempts") != 1 {
		t.Fatalf("duplicate task executed again: %v", row)
	}
	ds, e := k.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if e != nil || len(ds.Items) != 1 {
		t.Fatal("duplicate Kubernetes deployment")
	}
	deadline = time.Now().Add(3 * time.Second)
	for {
		if value, e := cache.Get(ctx, "instance_task:"+id).Result(); e == nil && value != "" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("task cache not updated")
		}
		time.Sleep(25 * time.Millisecond)
	}
}
