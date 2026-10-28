package cluster

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
)

type Client struct {
	K       kubernetes.Interface
	Dynamic dynamic.Interface
	REST    *rest.Config
	Config  config.Config
}

var Middleware = schema.GroupVersionResource{Group: "traefik.io", Version: "v1alpha1", Resource: "middlewares"}

func New(c config.Config) (*Client, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("Kubernetes config: %w", err)
	}
	cfg.Timeout = 30 * time.Second
	k, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	d, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{K: k, Dynamic: d, REST: cfg, Config: c}, nil
}
func Validate(v model.Record) error {
	for k, d := range (model.Record{"cpu_limit": 1.0, "memory_limit": "1Gi", "storage_limit": 1, "gpu_limit": 0, "gpu_memory_limit": 0, "namespace": "default", "replicas": 1}) {
		if v[k] == nil {
			v[k] = d
		}
	}
	cpu := model.Number(v, "cpu_limit")
	if cpu < 0.5 || cpu > 16 {
		return fmt.Errorf("CPU 必须为 0.5-16")
	}
	memory, err := resource.ParseQuantity(model.String(v, "memory_limit"))
	if err != nil || memory.Value() < 512<<20 || memory.Value() > 32<<30 {
		return fmt.Errorf("内存必须为 512Mi-32Gi")
	}
	for _, spec := range []struct {
		k        string
		min, max float64
	}{{"storage_limit", 1, 40}, {"gpu_limit", 0, 8}, {"gpu_memory_limit", 0, 1048576}, {"replicas", 1, 1}} {
		n := model.Number(v, spec.k)
		if n < spec.min || n > spec.max || n != float64(int(n)) {
			return fmt.Errorf("%s 无效", spec.k)
		}
	}
	return nil
}
func (c *Client) Provision(ctx context.Context, v model.Record) error {
	if c == nil {
		return fmt.Errorf("Kubernetes 未配置")
	}
	if err := Validate(v); err != nil {
		return err
	}
	id, ns := model.String(v, "id"), model.String(v, "namespace")
	name := "deploy-cs-" + id
	svc := "svc-cs-" + id
	pvc := "edu-pvc-cs-" + model.String(v, "user_id")
	cfg := model.Object(v["config"])
	password := model.String(cfg, "password")
	limits := corev1.ResourceList{corev1.ResourceCPU: resource.MustParse(strconv.FormatFloat(model.Number(v, "cpu_limit"), 'f', -1, 64)), corev1.ResourceMemory: resource.MustParse(model.String(v, "memory_limit")), corev1.ResourceEphemeralStorage: resource.MustParse(fmt.Sprintf("%dGi", int(model.Number(v, "storage_limit"))))}
	if gpu := model.Number(v, "gpu_limit"); gpu > 0 {
		limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprint(int(gpu)))
		if gm := model.Number(v, "gpu_memory_limit"); gm > 0 {
			limits["nvidia.com/gpumem"] = resource.MustParse(fmt.Sprint(int(gm)))
		}
	}
	cap, err := resource.ParseQuantity(c.Config.PVCCapacity)
	if err != nil {
		return err
	}
	_, err = c.K.CoreV1().PersistentVolumeClaims(ns).Create(ctx, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: pvc}, Spec: corev1.PersistentVolumeClaimSpec{AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(c.Config.PVCAccessMode)}, StorageClassName: &c.Config.StorageClass, Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: cap}}}}, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	image := model.String(v, "image")
	code := strings.Contains(strings.ToLower(image), "codeserver") || strings.Contains(strings.ToLower(image), "code-server")
	ports := []corev1.ContainerPort{{Name: "ssh", ContainerPort: 22}}
	servicePorts := []corev1.ServicePort{{Name: "ssh", Port: 80, TargetPort: intstr.FromInt32(22)}}
	env := []corev1.EnvVar{{Name: "TARGET_PASS", Value: password}, {Name: "ROOT_PASSWORD", Value: password}, {Name: "SSH_PASSWORD", Value: password}, {Name: "SSH_USERNAME", Value: model.String(v, "ssh_username")}}
	if code {
		ports = append(ports, corev1.ContainerPort{Name: "http", ContainerPort: 8080})
		servicePorts = []corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt32(8080)}, {Name: "ssh", Port: 22, TargetPort: intstr.FromInt32(22)}}
		env = append(env, corev1.EnvVar{Name: "CODE_SERVER_PASSWORD", Value: password}, corev1.EnvVar{Name: "PASSWORD", Value: password})
	}
	container := corev1.Container{Name: "ubuntu-runtime", Image: image, WorkingDir: "/workspace", Ports: ports, Env: env, Resources: corev1.ResourceRequirements{Limits: limits, Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("256Mi")}}, VolumeMounts: []corev1.VolumeMount{{Name: "project-data", MountPath: "/workspace"}}}
	if !code {
		container.Command = []string{"bash", "-lc", "mkdir -p /workspace /var/run/sshd; if ! command -v sshd >/dev/null; then apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y openssh-server; fi; printf 'root:%s\\n' \"$ROOT_PASSWORD\" | chpasswd; ssh-keygen -A; sed -i 's/^#\\?PermitRootLogin.*/PermitRootLogin yes/; s/^#\\?PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config; exec /usr/sbin/sshd -D -e"}
	}
	replicas := int32(1)
	labels := map[string]string{"app": name, "mixm/instance-id": id}
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}, Spec: appsv1.DeploymentSpec{Replicas: &replicas, Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}}, Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}, Spec: corev1.PodSpec{Containers: []corev1.Container{container}, Volumes: []corev1.Volume{{Name: "project-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvc}}}}}}}}
	_, err = c.K.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = c.K.CoreV1().Services(ns).Create(ctx, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: svc}, Spec: corev1.ServiceSpec{Selector: map[string]string{"app": name}, Ports: servicePorts}}, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	if code {
		path := "/instance/" + id
		mw := "mid-cs-" + id
		_, err = c.Dynamic.Resource(Middleware).Namespace(ns).Create(ctx, &unstructured.Unstructured{Object: map[string]any{"apiVersion": "traefik.io/v1alpha1", "kind": "Middleware", "metadata": map[string]any{"name": mw}, "spec": map[string]any{"stripPrefix": map[string]any{"prefixes": []any{path}, "forceSlash": true}}}}, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		class := "traefik"
		prefix := netv1.PathTypePrefix
		_, err = c.K.NetworkingV1().Ingresses(ns).Create(ctx, &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing-cs-" + id, Annotations: map[string]string{"traefik.ingress.kubernetes.io/router.middlewares": ns + "-" + mw + "@kubernetescrd"}}, Spec: netv1.IngressSpec{IngressClassName: &class, Rules: []netv1.IngressRule{{IngressRuleValue: netv1.IngressRuleValue{HTTP: &netv1.HTTPIngressRuleValue{Paths: []netv1.HTTPIngressPath{{Path: path, PathType: &prefix, Backend: netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: svc, Port: netv1.ServiceBackendPort{Number: 80}}}}}}}}}}}, metav1.CreateOptions{})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	longhornDone := true
	if sc, err := c.K.StorageV1().StorageClasses().Get(ctx, c.Config.StorageClass, metav1.GetOptions{}); err == nil && strings.Contains(strings.ToLower(sc.Provisioner), "longhorn") {
		longhornDone = false
	}
	return wait.PollUntilContextTimeout(ctx, 3*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		if !longhornDone {
			claim, err := c.K.CoreV1().PersistentVolumeClaims(ns).Get(ctx, pvc, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			if claim.Spec.VolumeName == "" {
				return false, nil
			}
			if err = c.setLonghornReplicas(ctx, claim.Spec.VolumeName); err != nil {
				return false, err
			}
			longhornDone = true
		}
		d, err := c.K.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return d.Status.ObservedGeneration >= d.Generation && d.Status.ReadyReplicas >= replicas && d.Status.AvailableReplicas >= replicas, nil
	})
}

func (c *Client) setLonghornReplicas(ctx context.Context, volume string) error {
	if c.Dynamic == nil {
		return fmt.Errorf("Longhorn requires dynamic client")
	}
	var last error
	for _, version := range []string{"v1beta2", "v1beta1"} {
		gvr := schema.GroupVersionResource{Group: "longhorn.io", Version: version, Resource: "volumes"}
		v, err := c.Dynamic.Resource(gvr).Namespace("longhorn-system").Get(ctx, volume, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				last = err
				continue
			}
			return err
		}
		spec := model.Object(v.Object["spec"])
		if model.Number(spec, "numberOfReplicas") == 1 {
			return nil
		}
		_, err = c.Dynamic.Resource(gvr).Namespace("longhorn-system").Patch(ctx, volume, types.MergePatchType, []byte(`{"spec":{"numberOfReplicas":1}}`), metav1.PatchOptions{})
		return err
	}
	return last
}
func (c *Client) Scale(ctx context.Context, ns, id string, n int32) error {
	if c == nil {
		return fmt.Errorf("Kubernetes 未配置")
	}
	d, err := c.K.AppsV1().Deployments(ns).GetScale(ctx, "deploy-cs-"+id, metav1.GetOptions{})
	if err != nil {
		return err
	}
	d.Spec.Replicas = n
	_, err = c.K.AppsV1().Deployments(ns).UpdateScale(ctx, d.Name, d, metav1.UpdateOptions{})
	return err
}
func (c *Client) Delete(ctx context.Context, ns, id string) error {
	if c == nil {
		return fmt.Errorf("Kubernetes 未配置")
	}
	policy := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &policy}
	for _, f := range []func() error{func() error { return c.K.AppsV1().Deployments(ns).Delete(ctx, "deploy-cs-"+id, opts) }, func() error { return c.K.CoreV1().Services(ns).Delete(ctx, "svc-cs-"+id, metav1.DeleteOptions{}) }, func() error {
		return c.K.NetworkingV1().Ingresses(ns).Delete(ctx, "ing-cs-"+id, metav1.DeleteOptions{})
	}, func() error {
		return c.Dynamic.Resource(Middleware).Namespace(ns).Delete(ctx, "mid-cs-"+id, metav1.DeleteOptions{})
	}} {
		if err := f(); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}
