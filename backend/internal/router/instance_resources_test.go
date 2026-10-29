package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

// TestInstanceResourcesRoundTrip 验证资源参数经过创建、持久化、接口查询和 Pod 配置后保持一致。
func TestInstanceResourcesRoundTrip(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user, err := repository.Write(ctx, db, "users", "", model.Record{
		"username":        "resource-admin",
		"hashed_password": "unused",
		"is_superuser":    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	settings := config.Config{
		Secret:        "resource-test-secret",
		SSHPort:       30022,
		CodeHost:      "localhost",
		PVCCapacity:   "10Gi",
		PVCAccessMode: "ReadWriteMany",
		StorageClass:  "local-path",
	}
	token, err := auth.Sign(settings.Secret, model.String(user, "username"), "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// 仅模拟集群资源创建与就绪状态，不连接业务 Kubernetes 集群。
	kubernetes := kubefake.NewSimpleClientset()
	kubernetes.PrependReactor("create", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
		deployment := action.(ktesting.CreateAction).GetObject().(*appsv1.Deployment)
		deployment.Status.ObservedGeneration = deployment.Generation
		deployment.Status.ReadyReplicas = 1
		deployment.Status.AvailableReplicas = 1
		return false, nil, nil
	})
	k := &cluster.Client{
		K:       kubernetes,
		Dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
		Config:  settings,
	}
	router := New(settings, db, nil, k)
	call := func(method, path string, payload any, status int) model.Record {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != status {
			t.Fatalf("%s %s returned %d, want %d", method, path, response.Code, status)
		}
		var envelope model.Record
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return model.Object(envelope["data"])
	}

	for _, input := range []model.Record{
		{
			"name":             "cpu",
			"cpu_limit":        0.5,
			"memory_limit":     "512Mi",
			"storage_limit":    5,
			"gpu_limit":        0,
			"gpu_memory_limit": 0,
		},
		{
			"name":             "gpu",
			"cpu_limit":        2.5,
			"memory_limit":     "3Gi",
			"storage_limit":    20,
			"gpu_limit":        2,
			"gpu_memory_limit": 3072,
		},
	} {
		t.Run(model.String(input, "name"), func(t *testing.T) {
			input["image"] = "mixm-codeserver:test"
			task := call(http.MethodPost, "/api/instances", input, http.StatusAccepted)
			id := model.String(task, "instance_id")
			stored, err := repository.Get(ctx, db, "instances", id)
			if err != nil {
				t.Fatal(err)
			}

			checkResources := func(record model.Record) {
				t.Helper()
				for _, key := range []string{"cpu_limit", "storage_limit", "gpu_limit", "gpu_memory_limit"} {
					if model.Number(record, key) != model.Number(input, key) {
						t.Errorf("resource %s changed: got %v, want %v", key, record[key], input[key])
					}
				}
				if record["memory_limit"] != input["memory_limit"] {
					t.Errorf("memory limit changed: got %v, want %v", record["memory_limit"], input["memory_limit"])
				}
			}
			checkResources(stored)
			checkResources(call(http.MethodGet, "/api/instances/"+id, nil, http.StatusOK))
			queued := call(http.MethodGet, "/api/instance-tasks/"+model.String(task, "task_id"), nil, http.StatusOK)
			checkResources(queued)
			for _, path := range []string{"/api/instances", "/api/users/me/instances"} {
				page := call(http.MethodGet, path, nil, http.StatusOK)
				found := false
				for _, record := range model.Objects(page["items"]) {
					if model.String(record, "id") == id {
						checkResources(record)
						found = true
					}
				}
				if !found {
					t.Fatalf("instance missing from %s", path)
				}
			}

			// 资源限制使用同一份数据库记录；临时存储与持久化工作卷分别校验。
			if err := k.Provision(ctx, stored); err != nil {
				t.Fatal(err)
			}
			deployment, err := kubernetes.AppsV1().Deployments("default").Get(ctx, "deploy-cs-"+id, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			limits := deployment.Spec.Template.Spec.Containers[0].Resources.Limits
			memory := resource.MustParse(model.String(input, "memory_limit"))
			if limits.Cpu().MilliValue() != int64(model.Number(input, "cpu_limit")*1000) ||
				limits.Memory().Value() != memory.Value() {
				t.Fatal("CPU or memory Pod limits do not match the instance")
			}
			if limits.StorageEphemeral().Value() != int64(model.Number(input, "storage_limit"))<<30 {
				t.Fatal("temporary storage Pod limit does not match the instance")
			}
			gpu, gpuMemory := limits["nvidia.com/gpu"], limits["nvidia.com/gpumem"]
			if gpu.Value() != int64(model.Number(input, "gpu_limit")) ||
				gpuMemory.Value() != int64(model.Number(input, "gpu_memory_limit")) {
				t.Fatal("GPU Pod limits do not match the instance")
			}
			claim, err := kubernetes.CoreV1().PersistentVolumeClaims("default").Get(ctx, "edu-pvc-cs-"+model.String(user, "id"), metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			capacity := claim.Spec.Resources.Requests[corev1.ResourceStorage]
			if capacity.Value() != 10<<30 {
				t.Fatal("temporary storage incorrectly changed persistent workspace capacity")
			}
		})
	}
}
