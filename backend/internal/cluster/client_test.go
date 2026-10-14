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
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

func TestResourceValidation(t *testing.T) {
	for _, v := range []store.Row{{"cpu_limit": 0.1}, {"cpu_limit": 17}, {"memory_limit": "garbage"}, {"memory_limit": "256Mi"}, {"memory_limit": "64Gi"}, {"storage_limit": 0}, {"gpu_limit": 9}, {"gpu_limit": 1.5}} {
		if e := Validate(v); e == nil {
			t.Fatalf("invalid resource accepted: %v", v)
		}
	}
	v := store.Row{}
	if e := Validate(v); e != nil {
		t.Fatal(e)
	}
	if v["memory_limit"] != "1Gi" {
		t.Fatal("defaults missing")
	}
}

func TestLonghornReplicaCompatibility(t *testing.T) {
	volume := &unstructured.Unstructured{Object: map[string]any{"apiVersion": "longhorn.io/v1beta2", "kind": "Volume", "metadata": map[string]any{"name": "test-volume", "namespace": "longhorn-system"}, "spec": map[string]any{"numberOfReplicas": int64(3)}}}
	dynamic := fake.NewSimpleDynamicClient(runtime.NewScheme(), volume)
	c := &Client{Dynamic: dynamic}
	if e := c.setLonghornReplicas(context.Background(), "test-volume"); e != nil {
		t.Fatal(e)
	}
	v, e := dynamic.Resource(volume.GroupVersionKind().GroupVersion().WithResource("volumes")).Namespace("longhorn-system").Get(context.Background(), "test-volume", metav1.GetOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if store.Number(store.Object(v.Object["spec"]), "numberOfReplicas") != 1 {
		t.Fatal("legacy Longhorn replica configuration changed")
	}
}
func TestProvisionIsIdempotentAndPreservesUserPVC(t *testing.T) {
	k := kubefake.NewSimpleClientset()
	k.PrependReactor("get", "deployments", func(a ktesting.Action) (bool, runtime.Object, error) {
		get := a.(ktesting.GetAction)
		obj, e := k.Tracker().Get(appsv1.SchemeGroupVersion.WithResource("deployments"), a.GetNamespace(), get.GetName())
		if e != nil {
			return true, nil, e
		}
		d := obj.(*appsv1.Deployment).DeepCopy()
		d.Status.ObservedGeneration = d.Generation
		d.Status.ReadyReplicas = 1
		d.Status.AvailableReplicas = 1
		return true, d, nil
	})
	c := &Client{K: k, Dynamic: fake.NewSimpleDynamicClient(runtime.NewScheme()), Config: config.Config{PVCCapacity: "10Gi", PVCAccessMode: "ReadWriteMany", StorageClass: "longhorn"}}
	v := store.Row{"id": "test-id", "user_id": "student-id", "image": "example/code-server:latest", "namespace": "default", "config": `{"password":"secret"}`}
	for i := 0; i < 2; i++ {
		if e := c.Provision(context.Background(), v); e != nil {
			t.Fatal(e)
		}
	}
	ds, e := k.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	if e != nil || len(ds.Items) != 1 {
		t.Fatalf("duplicate deployments: %v", e)
	}
	d := ds.Items[0]
	if d.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "edu-pvc-cs-student-id" {
		t.Fatal("user volume not reused")
	}
	if e = c.Delete(context.Background(), "default", "test-id"); e != nil {
		t.Fatal(e)
	}
	if _, e = k.CoreV1().PersistentVolumeClaims("default").Get(context.Background(), "edu-pvc-cs-student-id", metav1.GetOptions{}); e != nil {
		t.Fatal("user PVC deleted", e)
	}
}
