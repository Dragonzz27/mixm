package cluster

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
)

func TestResourceValidation(t *testing.T) {
	for _, v := range []model.Record{{"cpu_limit": 0.1}, {"cpu_limit": 17}, {"memory_limit": "garbage"}, {"memory_limit": "256Mi"}, {"memory_limit": "64Gi"}, {"storage_limit": 0}, {"gpu_limit": 9}, {"gpu_limit": 1.5}} {
		if err := Validate(v); err == nil {
			t.Fatalf("invalid resource accepted: %v", v)
		}
	}
	v := model.Record{}
	if err := Validate(v); err != nil {
		t.Fatal(err)
	}
	if v["memory_limit"] != "1Gi" {
		t.Fatal("defaults missing")
	}
}

func TestLonghornReplicaConfiguration(t *testing.T) {
	volume := &unstructured.Unstructured{Object: map[string]any{"apiVersion": "longhorn.io/v1beta2", "kind": "Volume", "metadata": map[string]any{"name": "test-volume", "namespace": "longhorn-system"}, "spec": map[string]any{"numberOfReplicas": int64(3)}}}
	dynamic := fake.NewSimpleDynamicClient(runtime.NewScheme(), volume)
	c := &Client{Dynamic: dynamic}
	if err := c.setLonghornReplicas(context.Background(), "test-volume"); err != nil {
		t.Fatal(err)
	}
	v, err := dynamic.Resource(volume.GroupVersionKind().GroupVersion().WithResource("volumes")).Namespace("longhorn-system").Get(context.Background(), "test-volume", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if model.Number(model.Object(v.Object["spec"]), "numberOfReplicas") != 1 {
		t.Fatal("Longhorn replica configuration was not applied")
	}
}
func TestProvisionIsIdempotentAndPreservesUserPVC(t *testing.T) {
	k := kubefake.NewSimpleClientset()
	k.PrependReactor("get", "deployments", func(a ktesting.Action) (bool, runtime.Object, error) {
		get := a.(ktesting.GetAction)
		obj, err := k.Tracker().Get(appsv1.SchemeGroupVersion.WithResource("deployments"), a.GetNamespace(), get.GetName())
		if err != nil {
			return true, nil, err
		}
		d := obj.(*appsv1.Deployment).DeepCopy()
		d.Status.ObservedGeneration = d.Generation
		d.Status.ReadyReplicas = 1
		d.Status.AvailableReplicas = 1
		return true, d, nil
	})
	c := &Client{K: k, Dynamic: fake.NewSimpleDynamicClient(runtime.NewScheme()), Config: config.Config{PVCCapacity: "10Gi", PVCAccessMode: "ReadWriteMany", StorageClass: "longhorn"}}
	v := model.Record{"id": "test-id", "user_id": "student-id", "image": "example/code-server:latest", "namespace": "default", "config": `{"password":"secret"}`}
	for i := 0; i < 2; i++ {
		if err := c.Provision(context.Background(), v); err != nil {
			t.Fatal(err)
		}
	}
	ds, err := k.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	if err != nil || len(ds.Items) != 1 {
		t.Fatalf("duplicate deployments: %v", err)
	}
	d := ds.Items[0]
	if d.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "edu-pvc-cs-student-id" {
		t.Fatal("user volume not reused")
	}
	if err = c.Delete(context.Background(), "default", "test-id"); err != nil {
		t.Fatal(err)
	}
	if _, err = k.CoreV1().PersistentVolumeClaims("default").Get(context.Background(), "edu-pvc-cs-student-id", metav1.GetOptions{}); err != nil {
		t.Fatal("user PVC deleted", err)
	}
}
