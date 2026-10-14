package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"mixm-edu/backend/internal/store"
)

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
func (s *Server) clusterRoute(pattern string, h handler) {
	s.route(pattern, func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Clusters"); e != nil {
			return nil, e
		}
		if s.Cluster == nil {
			return nil, fail(503, "Kubernetes 未配置")
		}
		if r.Method == "DELETE" && r.PathValue("namespace") == "kube-system" {
			return nil, fail(403, "不能删除系统命名空间资源")
		}
		v, e := h(w, r)
		if e != nil {
			return nil, e
		}
		if r.Method == "DELETE" {
			respond(w, 200, v)
			return nil, nil
		}
		return v, nil
	})
}
func (s *Server) clusterRoutes() {
	s.clusterRoute("GET /api/clusters/pods", func(w http.ResponseWriter, r *http.Request) (any, error) {
		pods, e := s.Cluster.K.CoreV1().Pods(r.URL.Query().Get("namespace")).List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
		}
		out := []store.Row{}
		for _, p := range pods.Items {
			containers := []string{}
			for _, c := range p.Spec.Containers {
				containers = append(containers, c.Name)
			}
			restarts := int32(0)
			for _, c := range p.Status.ContainerStatuses {
				restarts += c.RestartCount
			}
			out = append(out, store.Row{"name": p.Name, "namespace": p.Namespace, "node": p.Spec.NodeName, "status": strings.ToLower(string(p.Status.Phase)), "containers": containers, "restarts": restarts, "age": age(p.CreationTimestamp)})
		}
		return store.Row{"total": len(out), "pods": out}, nil
	})
	s.clusterRoute("GET /api/clusters/nodes", func(w http.ResponseWriter, r *http.Request) (any, error) {
		nodes, e := s.Cluster.K.CoreV1().Nodes().List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
		}
		pods, e := s.Cluster.K.CoreV1().Pods("").List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
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
			metrics, err := s.Cluster.Dynamic.Resource(schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "nodes"}).List(r.Context(), metav1.ListOptions{})
			if err == nil {
				for _, item := range metrics.Items {
					usage := store.Object(item.Object["usage"])
					if q, err := resource.ParseQuantity(store.String(usage, "cpu")); err == nil {
						cpuUsage[item.GetName()] = q.MilliValue()
					}
					if q, err := resource.ParseQuantity(store.String(usage, "memory")); err == nil {
						memoryUsage[item.GetName()] = q.Value()
					}
				}
			}
		}
		out := []store.Row{}
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
			out = append(out, store.Row{"name": n.Name, "status": status, "roles": roles, "ip": ip, "version": n.Status.NodeInfo.KubeletVersion, "age": age(n.CreationTimestamp), "cpu_capacity": n.Status.Capacity.Cpu().MilliValue(), "memory_capacity": n.Status.Capacity.Memory().Value(), "cpu_usage": cpuUsage[n.Name], "memory_usage": memoryUsage[n.Name], "pods_count": counts[n.Name]})
		}
		return store.Row{"total": len(out), "nodes": out}, nil
	})
	s.clusterRoute("GET /api/clusters/deployments", func(w http.ResponseWriter, r *http.Request) (any, error) {
		ds, e := s.Cluster.K.AppsV1().Deployments("").List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
		}
		out := []store.Row{}
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
			out = append(out, store.Row{"name": d.Name, "namespace": d.Namespace, "replicas": replicas, "available_replicas": d.Status.AvailableReplicas, "status": status, "age": age(d.CreationTimestamp), "labels": d.Labels})
		}
		return store.Row{"total": len(out), "deployments": out}, nil
	})
	s.clusterRoute("DELETE /api/clusters/deployments/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return store.Row{"success": true, "message": "部署已删除"}, s.Cluster.K.AppsV1().Deployments(r.PathValue("namespace")).Delete(r.Context(), r.PathValue("name"), metav1.DeleteOptions{})
	})
	s.clusterRoute("DELETE /api/clusters/pods/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return store.Row{"success": true, "message": "Pod 已删除"}, s.Cluster.K.CoreV1().Pods(r.PathValue("namespace")).Delete(r.Context(), r.PathValue("name"), metav1.DeleteOptions{})
	})
	s.clusterRoute("GET /api/clusters/pods/{namespace}/{name}/logs", func(w http.ResponseWriter, r *http.Request) (any, error) {
		tail := int64(100)
		opts := &corev1.PodLogOptions{Container: r.URL.Query().Get("container"), TailLines: &tail}
		if _, e := fmt.Sscan(r.URL.Query().Get("tail_lines"), &tail); e == nil && (tail < 1 || tail > 10000) {
			return nil, fail(422, "tail_lines 无效")
		}
		logs, e := s.Cluster.K.CoreV1().Pods(r.PathValue("namespace")).GetLogs(r.PathValue("name"), opts).DoRaw(r.Context())
		return store.Row{"success": true, "logs": string(logs), "container": opts.Container, "pod_name": r.PathValue("name"), "namespace": r.PathValue("namespace")}, e
	})
	s.clusterRoute("GET /api/clusters/pods/{namespace}/{name}/describe", func(w http.ResponseWriter, r *http.Request) (any, error) {
		p, e := s.Cluster.K.CoreV1().Pods(r.PathValue("namespace")).Get(r.Context(), r.PathValue("name"), metav1.GetOptions{})
		if e != nil {
			return nil, e
		}
		containers := []store.Row{}
		for _, c := range p.Spec.Containers {
			containers = append(containers, store.Row{"name": c.Name, "image": c.Image, "image_pull_policy": c.ImagePullPolicy, "command": strings.Join(c.Command, " "), "args": strings.Join(c.Args, " "), "ports": c.Ports, "resources": c.Resources})
		}
		conds := []store.Row{}
		for _, c := range p.Status.Conditions {
			conds = append(conds, store.Row{"type": c.Type, "status": c.Status, "reason": c.Reason, "message": c.Message, "last_transition_time": c.LastTransitionTime})
		}
		events, e := s.Cluster.K.CoreV1().Events(p.Namespace).List(r.Context(), metav1.ListOptions{FieldSelector: "involvedObject.name=" + p.Name})
		if e != nil {
			return nil, e
		}
		evs := []store.Row{}
		for _, ev := range events.Items {
			evs = append(evs, store.Row{"type": ev.Type, "reason": ev.Reason, "message": ev.Message, "source": ev.Source.Component, "count": ev.Count, "first_timestamp": ev.FirstTimestamp, "last_timestamp": ev.LastTimestamp})
		}
		return store.Row{"success": true, "pod": store.Row{"name": p.Name, "namespace": p.Namespace, "node": p.Spec.NodeName, "status": p.Status.Phase, "ip": p.Status.PodIP, "host_ip": p.Status.HostIP, "start_time": p.Status.StartTime, "labels": p.Labels, "annotations": p.Annotations, "containers": containers, "conditions": conds, "events": evs}}, nil
	})
	s.clusterRoute("GET /api/clusters/pv", func(w http.ResponseWriter, r *http.Request) (any, error) {
		pvs, e := s.Cluster.K.CoreV1().PersistentVolumes().List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
		}
		out := []store.Row{}
		for _, v := range pvs.Items {
			claim := ""
			if v.Spec.ClaimRef != nil {
				claim = v.Spec.ClaimRef.Namespace + "/" + v.Spec.ClaimRef.Name
			}
			out = append(out, store.Row{"name": v.Name, "capacity": v.Spec.Capacity, "access_modes": v.Spec.AccessModes, "reclaim_policy": v.Spec.PersistentVolumeReclaimPolicy, "status": v.Status.Phase, "claim": claim, "storage_class": v.Spec.StorageClassName, "age": age(v.CreationTimestamp)})
		}
		return store.Row{"total": len(out), "pvs": out}, nil
	})
	s.clusterRoute("DELETE /api/clusters/pv/{name}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.Cluster.K.CoreV1().PersistentVolumes().Get(r.Context(), r.PathValue("name"), metav1.GetOptions{})
		if e != nil {
			return nil, e
		}
		if v.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimRetain || v.Status.Phase == corev1.VolumeBound {
			return nil, fail(409, "请先处理 Retain 数据或解除 PVC 绑定")
		}
		return store.Row{"success": true, "message": "PV 已删除"}, s.Cluster.K.CoreV1().PersistentVolumes().Delete(r.Context(), v.Name, metav1.DeleteOptions{})
	})
	s.clusterRoute("GET /api/clusters/pvc", func(w http.ResponseWriter, r *http.Request) (any, error) {
		pvcs, e := s.Cluster.K.CoreV1().PersistentVolumeClaims(r.URL.Query().Get("namespace")).List(r.Context(), metav1.ListOptions{})
		if e != nil {
			return nil, e
		}
		out := []store.Row{}
		for _, v := range pvcs.Items {
			out = append(out, store.Row{"name": v.Name, "namespace": v.Namespace, "storage_class": v.Spec.StorageClassName, "capacity": v.Status.Capacity, "access_modes": v.Spec.AccessModes, "volume_mode": v.Spec.VolumeMode, "status": v.Status.Phase, "volume": v.Spec.VolumeName, "age": age(v.CreationTimestamp)})
		}
		return store.Row{"total": len(out), "pvcs": out}, nil
	})
	s.clusterRoute("DELETE /api/clusters/pvc/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		return store.Row{"success": true, "message": "PVC 已删除"}, s.Cluster.K.CoreV1().PersistentVolumeClaims(r.PathValue("namespace")).Delete(r.Context(), r.PathValue("name"), metav1.DeleteOptions{})
	})
}
