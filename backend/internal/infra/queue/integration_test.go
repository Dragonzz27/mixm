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

	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

// TestRabbitMQOutboxAndDuplicateDelivery 验证任务可靠投递与重复消息处理。
func TestRabbitMQOutboxAndDuplicateDelivery(t *testing.T) {
	if os.Getenv("TEST_RABBITMQ_URL") == "" || os.Getenv("TEST_REDIS_URL") == "" {
		t.Skip("set TEST_RABBITMQ_URL and TEST_REDIS_URL for queue integration")
	}

	db := testutil.Database(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts, err := redis.ParseURL(os.Getenv("TEST_REDIS_URL"))
	if err != nil {
		t.Fatal(err)
	}

	cache := redis.NewClient(opts)
	defer cache.Close()

	name := "test_" + uuid.NewString()
	cfg := config.Config{
		RabbitURL:     os.Getenv("TEST_RABBITMQ_URL"),
		Queue:         name,
		PVCCapacity:   "10Gi",
		PVCAccessMode: "ReadWriteMany",
		StorageClass:  "longhorn",
	}
	conn, ch, err := channel(cfg)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	defer ch.Close()
	defer ch.QueueDelete(name, false, false, false)

	u, err := repository.Write(ctx, db, "users", "", model.Record{"username": "queue-test", "hashed_password": "unused"})
	if err != nil {
		t.Fatal(err)
	}

	v := model.Record{
		"name":    "test",
		"user_id": u["id"],
		"image":   "ubuntu:24.04",
		"config":  `{"password":"secret"}`,
	}
	if err = cluster.Validate(v); err != nil {
		t.Fatal(err)
	}

	inst, err := repository.Write(ctx, db, "instances", "", v)
	if err != nil {
		t.Fatal(err)
	}

	task, err := repository.Write(
		ctx,
		db,
		"instance_tasks",
		"",
		model.Record{
			"user_id":     u["id"],
			"instance_id": inst["id"],
			"payload":     v,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	id := model.String(task, "id")
	defer cache.Del(context.Background(), "instance_task:"+id)

	k := kubefake.NewSimpleClientset()
	k.PrependReactor(
		"get",
		"deployments",
		func(a ktesting.Action) (bool, runtime.Object, error) {
			obj, err := k.Tracker().Get(
				appsv1.SchemeGroupVersion.WithResource("deployments"),
				a.GetNamespace(),
				a.(ktesting.GetAction).GetName(),
			)
			if err != nil {
				return true, nil, err
			}

			d := obj.(*appsv1.Deployment).DeepCopy()
			d.Status.ObservedGeneration = d.Generation
			d.Status.AvailableReplicas = 1
			d.Status.ReadyReplicas = 1

			return true, d, nil
		},
	)
	client := &cluster.Client{
		K:       k,
		Dynamic: fake.NewSimpleDynamicClient(runtime.NewScheme()),
		Config:  cfg,
	}
	publisherDone, workerDone := make(chan struct{}), make(chan error, 1)
	go func() {
		Dispatch(ctx, cfg, db)
		close(publisherDone)
	}()
	go func() {
		workerDone <- Work(ctx, cfg, db, cache, client)
	}()
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
		row, err := repository.Get(ctx, db, "instance_tasks", id)
		if err != nil {
			t.Fatal(err)
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

	if err = ch.PublishWithContext(ctx, "", name, false, false, amqp.Publishing{Body: []byte(id), DeliveryMode: amqp.Persistent}); err != nil {
		t.Fatal(err)
	}

	if err = process(ctx, db, cache, client, id); err != nil {
		t.Fatal(err)
	}

	row, err := repository.Get(ctx, db, "instance_tasks", id)
	if err != nil || model.Number(row, "attempts") != 1 {
		t.Fatalf("duplicate task executed again: %v", row)
	}

	ds, err := k.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil || len(ds.Items) != 1 {
		t.Fatal("duplicate Kubernetes deployment")
	}

	deadline = time.Now().Add(3 * time.Second)

	for {
		if value, err := cache.Get(ctx, "instance_task:"+id).Result(); err == nil && value != "" {
			break
		}

		if time.Now().After(deadline) {
			t.Fatal("task cache not updated")
		}

		time.Sleep(25 * time.Millisecond)
	}
}
