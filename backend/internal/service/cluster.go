package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"mixm/backend/internal/model"
)

type Cluster struct{ *Core }

// age 将资源创建时间转换为可读的持续时长。
func age(t metav1.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	d := time.Since(t.Time)
	if d >= 24*time.Hour {
		return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
	}

	if d >= time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}

	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// GetPods 读取集群 Pod 列表及运行状态。
func (s *Cluster) GetPods(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	pods, err := s.Cluster.K.CoreV1().Pods(query.Get("namespace")).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := []model.Record{}

	for _, p := range pods.Items {
		containers := []string{}

		for _, c := range p.Spec.Containers {
			containers = append(containers, c.Name)
		}

		restarts := int32(0)

		for _, c := range p.Status.ContainerStatuses {
			restarts += c.RestartCount
		}

		out = append(
			out,
			model.Record{
				"name":       p.Name,
				"namespace":  p.Namespace,
				"node":       p.Spec.NodeName,
				"status":     strings.ToLower(string(p.Status.Phase)),
				"containers": containers,
				"restarts":   restarts,
				"age":        age(p.CreationTimestamp),
			},
		)
	}

	return model.Record{"total": len(out), "pods": out}, nil
}

// GetNodes 汇总集群节点状态与资源使用信息。
func (s *Cluster) GetNodes(ctx context.Context, actor model.Actor) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	nodes, err := s.Cluster.K.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := s.Cluster.K.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	cpuUsage, memoryUsage := map[string]int64{}, map[string]int64{}

	for _, p := range pods.Items {
		counts[p.Spec.NodeName]++

		for _, c := range p.Spec.Containers {
			cpuUsage[p.Spec.NodeName] += c.Resources.Requests.Cpu().MilliValue()
			memoryUsage[p.Spec.NodeName] += c.Resources.Requests.Memory().Value()
		}
	}

	if s.Cluster.Dynamic != nil {
		metrics, err := s.Cluster.Dynamic.Resource(schema.GroupVersionResource{
			Group:    "metrics.k8s.io",
			Version:  "v1beta1",
			Resource: "nodes",
		}).List(
			ctx,
			metav1.ListOptions{},
		)
		if err == nil {
			for _, item := range metrics.Items {
				usage := model.Object(item.Object["usage"])
				if q, err := resource.ParseQuantity(model.String(usage, "cpu")); err == nil {
					cpuUsage[item.GetName()] = q.MilliValue()
				}

				if q, err := resource.ParseQuantity(model.String(usage, "memory")); err == nil {
					memoryUsage[item.GetName()] = q.Value()
				}
			}
		}
	}

	out := []model.Record{}

	for _, n := range nodes.Items {
		status := "NotReady"

		for _, cond := range n.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
			}
		}

		ip := ""

		for _, a := range n.Status.Addresses {
			if a.Type == corev1.NodeInternalIP {
				ip = a.Address
			}
		}

		roles := []string{}

		for k := range n.Labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") {
				roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
			}
		}

		out = append(
			out,
			model.Record{
				"name":            n.Name,
				"status":          status,
				"roles":           roles,
				"ip":              ip,
				"version":         n.Status.NodeInfo.KubeletVersion,
				"age":             age(n.CreationTimestamp),
				"cpu_capacity":    n.Status.Capacity.Cpu().MilliValue(),
				"memory_capacity": n.Status.Capacity.Memory().Value(),
				"cpu_usage":       cpuUsage[n.Name],
				"memory_usage":    memoryUsage[n.Name],
				"pods_count":      counts[n.Name],
			},
		)
	}

	return model.Record{"total": len(out), "nodes": out}, nil
}

// GetDeployments 读取 Deployment 副本数和运行状态。
func (s *Cluster) GetDeployments(ctx context.Context, actor model.Actor) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	ds, err := s.Cluster.K.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := []model.Record{}

	for _, d := range ds.Items {
		replicas := int32(1)
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}

		status := "pending"
		if replicas == 0 {
			status = "stopped"
		} else if d.Status.AvailableReplicas >= replicas {
			status = "running"
		}

		out = append(
			out,
			model.Record{
				"name":               d.Name,
				"namespace":          d.Namespace,
				"replicas":           replicas,
				"available_replicas": d.Status.AvailableReplicas,
				"status":             status,
				"age":                age(d.CreationTimestamp),
				"labels":             d.Labels,
			},
		)
	}

	return model.Record{"total": len(out), "deployments": out}, nil
}

// DeleteDeployments 删除指定 Deployment 并保护系统资源。
func (s *Cluster) DeleteDeployments(ctx context.Context, actor model.Actor, namespace string, name string) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	if namespace == "kube-system" {
		return nil, fail(http.StatusForbidden, "不能删除系统命名空间资源")
	}

	return model.Record{"success": true, "message": "部署已删除"}, s.Cluster.K.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// DeletePods 删除指定 Pod 并保护系统资源。
func (s *Cluster) DeletePods(ctx context.Context, actor model.Actor, namespace string, name string) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	if namespace == "kube-system" {
		return nil, fail(http.StatusForbidden, "不能删除系统命名空间资源")
	}

	return model.Record{"success": true, "message": "Pod 已删除"}, s.Cluster.K.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetPodsLogs 读取指定 Pod 容器的最近日志。
func (s *Cluster) GetPodsLogs(
	ctx context.Context,
	actor model.Actor,
	query url.Values,
	namespace string,
	name string,
) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	tail := int64(100)
	opts := &corev1.PodLogOptions{Container: query.Get("container"), TailLines: &tail}
	if _, err := fmt.Sscan(query.Get("tail_lines"), &tail); err == nil && (tail < 1 || tail > 10000) {
		return nil, fail(http.StatusUnprocessableEntity, "tail_lines 无效")
	}

	logs, err := s.Cluster.K.CoreV1().Pods(namespace).GetLogs(name, opts).DoRaw(ctx)

	return model.Record{
		"success":   true,
		"logs":      string(logs),
		"container": opts.Container,
		"pod_name":  name,
		"namespace": namespace,
	}, err
}

// GetPodsDescribe 汇总 Pod 状态、容器配置及事件。
func (s *Cluster) GetPodsDescribe(ctx context.Context, actor model.Actor, namespace string, name string) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	p, err := s.Cluster.K.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	containers := []model.Record{}

	for _, c := range p.Spec.Containers {
		containers = append(
			containers,
			model.Record{
				"name":              c.Name,
				"image":             c.Image,
				"image_pull_policy": c.ImagePullPolicy,
				"command":           strings.Join(c.Command, " "),
				"args":              strings.Join(c.Args, " "),
				"ports":             c.Ports,
				"resources":         c.Resources,
			},
		)
	}

	conds := []model.Record{}

	for _, c := range p.Status.Conditions {
		conds = append(
			conds,
			model.Record{
				"type":                 c.Type,
				"status":               c.Status,
				"reason":               c.Reason,
				"message":              c.Message,
				"last_transition_time": c.LastTransitionTime,
			},
		)
	}

	events, err := s.Cluster.K.CoreV1().Events(p.Namespace).List(ctx, metav1.ListOptions{FieldSelector: "involvedObject.name=" + p.Name})
	if err != nil {
		return nil, err
	}

	evs := []model.Record{}

	for _, ev := range events.Items {
		evs = append(
			evs,
			model.Record{
				"type":            ev.Type,
				"reason":          ev.Reason,
				"message":         ev.Message,
				"source":          ev.Source.Component,
				"count":           ev.Count,
				"first_timestamp": ev.FirstTimestamp,
				"last_timestamp":  ev.LastTimestamp,
			},
		)
	}

	return model.Record{
		"success": true,
		"pod": model.Record{
			"name":        p.Name,
			"namespace":   p.Namespace,
			"node":        p.Spec.NodeName,
			"status":      p.Status.Phase,
			"ip":          p.Status.PodIP,
			"host_ip":     p.Status.HostIP,
			"start_time":  p.Status.StartTime,
			"labels":      p.Labels,
			"annotations": p.Annotations,
			"containers":  containers,
			"conditions":  conds,
			"events":      evs,
		},
	}, nil
}

// GetPv 读取持久卷列表及绑定状态。
func (s *Cluster) GetPv(ctx context.Context, actor model.Actor) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	pvs, err := s.Cluster.K.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := []model.Record{}

	for _, v := range pvs.Items {
		claim := ""
		if v.Spec.ClaimRef != nil {
			claim = v.Spec.ClaimRef.Namespace + "/" + v.Spec.ClaimRef.Name
		}

		out = append(
			out,
			model.Record{
				"name":           v.Name,
				"capacity":       v.Spec.Capacity,
				"access_modes":   v.Spec.AccessModes,
				"reclaim_policy": v.Spec.PersistentVolumeReclaimPolicy,
				"status":         v.Status.Phase,
				"claim":          claim,
				"storage_class":  v.Spec.StorageClassName,
				"age":            age(v.CreationTimestamp),
			},
		)
	}

	return model.Record{"total": len(out), "pvs": out}, nil
}

// DeletePv 检查绑定和回收策略后删除持久卷。
func (s *Cluster) DeletePv(ctx context.Context, actor model.Actor, name string) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	v, err := s.Cluster.K.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if v.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimRetain || v.Status.Phase == corev1.VolumeBound {
		return nil, fail(http.StatusConflict, "请先处理 Retain 数据或解除 PVC 绑定")
	}

	return model.Record{"success": true, "message": "PV 已删除"}, s.Cluster.K.CoreV1().PersistentVolumes().Delete(ctx, v.Name, metav1.DeleteOptions{})
}

// GetPvc 读取指定范围内的持久卷声明。
func (s *Cluster) GetPvc(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	pvcs, err := s.Cluster.K.CoreV1().PersistentVolumeClaims(query.Get("namespace")).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := []model.Record{}

	for _, v := range pvcs.Items {
		out = append(
			out,
			model.Record{
				"name":          v.Name,
				"namespace":     v.Namespace,
				"storage_class": v.Spec.StorageClassName,
				"capacity":      v.Status.Capacity,
				"access_modes":  v.Spec.AccessModes,
				"volume_mode":   v.Spec.VolumeMode,
				"status":        v.Status.Phase,
				"volume":        v.Spec.VolumeName,
				"age":           age(v.CreationTimestamp),
			},
		)
	}

	return model.Record{"total": len(out), "pvcs": out}, nil
}

// DeletePvc 删除指定持久卷声明并保护系统资源。
func (s *Cluster) DeletePvc(ctx context.Context, actor model.Actor, namespace string, name string) (any, error) {
	if err := s.permission(ctx, actor, "Clusters"); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	if namespace == "kube-system" {
		return nil, fail(http.StatusForbidden, "不能删除系统命名空间资源")
	}

	return model.Record{"success": true, "message": "PVC 已删除"}, s.Cluster.K.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
