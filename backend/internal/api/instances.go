package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/resource"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/store"
)

func (s *Server) ownInstance(r *http.Request, id string) (store.Row, error) {
	v, e := store.Get(r.Context(), s.DB, "instances", id)
	if e != nil {
		return nil, e
	}
	if v["user_id"] != user(r)["id"] {
		if e = s.permission(r, "Instances"); e != nil {
			return nil, e
		}
	}
	return v, nil
}
func (s *Server) instanceResponse(v store.Row) store.Row {
	id := store.String(v, "id")
	v["endpoint"] = "/instance/" + id
	cfg := store.Object(v["config"])
	v["ssh_host"], v["ssh_port"] = s.Config.SSHHost, s.Config.SSHPort
	v["ssh_password"] = cfg["password"]
	v["hostIp"] = s.Config.CodeHost
	v["code_server_host"] = s.Config.CodeHost
	v["ssh_command"] = fmt.Sprintf("ssh %s@%s -p %d", store.String(v, "ssh_username"), s.Config.SSHHost, s.Config.SSHPort)
	v["ssh_connection"] = v["ssh_command"]
	v["access_url"] = "http://" + s.Config.CodeHost + "/instance/" + id + "/"
	v["code_server_url"] = v["access_url"]
	v["password"] = cfg["password"]
	return v
}
func taskResponse(t store.Row) store.Row {
	v := store.Object(t["payload"])
	for _, k := range []string{"user_id", "status", "instance_id", "error", "created_at", "started_at", "completed_at"} {
		v[k] = t[k]
	}
	v["task_id"] = t["id"]
	v["instance_name"] = v["name"]
	return v
}
func (s *Server) enqueue(r *http.Request, v store.Row) (store.Row, error) {
	if s.Cluster == nil {
		return nil, fail(503, "Kubernetes 未配置")
	}
	uid := store.String(v, "user_id")
	if uid == "" {
		uid = store.String(user(r), "id")
	}
	if uid != store.String(user(r), "id") {
		if e := s.permission(r, "Instances"); e != nil {
			return nil, e
		}
	}
	u, e := store.Get(r.Context(), s.DB, "users", uid)
	if e != nil {
		return nil, e
	}
	if !store.Bool(u, "is_active") {
		return nil, fail(422, "用户已停用")
	}
	v["user_id"] = uid
	if e = cluster.Validate(v); e != nil {
		return nil, fail(422, e.Error())
	}
	if store.String(v, "repo_id") != "" {
		repo, e := store.Get(r.Context(), s.DB, "image_repositories", store.String(v, "repo_id"))
		if e != nil {
			return nil, e
		}
		if !store.Bool(repo, "is_public") && repo["user_id"] != uid && !store.Bool(user(r), "is_superuser") {
			return nil, fail(403, "无权使用镜像")
		}
		if e = require(v, "repo_tag"); e != nil {
			return nil, e
		}
		v["image"] = strings.TrimSuffix(s.Config.RegistryPrefix, "/") + "/" + store.String(repo, "name") + ":" + store.String(v, "repo_tag")
	}
	if store.String(v, "image") == "" {
		v["image"] = "ubuntu:24.04"
	}
	if store.String(v, "name") == "" {
		v["name"] = "instance"
	}
	v["name"] = fmt.Sprintf("%s-%d", v["name"], time.Now().UnixNano())
	id := uuid.NewString()
	v["ssh_username"] = store.String(u, "username") + strings.ReplaceAll(id, "-", "")[:8]
	v["ssh_port"] = s.Config.SSHPort
	v["status"] = "pending"
	b := make([]byte, 24)
	if _, e = rand.Read(b); e != nil {
		return nil, e
	}
	v["config"] = store.JSON(store.Row{"password": hex.EncodeToString(b), "working_dir": "/workspace", "image": v["image"]})
	memory, _ := resource.ParseQuantity(store.String(v, "memory_limit"))
	v["memory_limit_mb"] = memory.Value() / (1 << 20)
	tx, e := s.DB.Begin(r.Context())
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(r.Context())
	instance, e := store.Write(r.Context(), tx, "instances", "", v)
	if e != nil {
		return nil, e
	}
	task, e := store.Write(r.Context(), tx, "instance_tasks", "", store.Row{"user_id": uid, "instance_id": instance["id"], "payload": v})
	if e != nil {
		return nil, e
	}
	if e = tx.Commit(r.Context()); e != nil {
		return nil, e
	}
	return store.Row{"task_id": task["id"], "instance_id": instance["id"], "instance_name": instance["name"], "user_id": uid, "status": "pending", "message": "实例已加入创建队列", "config": store.Row{"repo_id": v["repo_id"], "repo_tag": v["repo_tag"], "cpu_limit": v["cpu_limit"], "memory_limit": v["memory_limit"], "storage_limit": v["storage_limit"], "gpu_limit": v["gpu_limit"], "gpu_memory_limit": v["gpu_memory_limit"]}}, nil
}
func (s *Server) instanceRoutes() {
	for _, p := range []string{"/api/instances", "/api/instances/start"} {
		s.route("POST "+p, func(w http.ResponseWriter, r *http.Request) (any, error) {
			v, e := body(r)
			if e != nil {
				return nil, e
			}
			out, e := s.enqueue(r, v)
			if e != nil {
				return nil, e
			}
			respond(w, 202, out)
			return nil, nil
		})
	}
	for _, p := range []string{"/api/instances", "/api/instances/me", "/api/instances/user/{user_id}"} {
		s.route("GET "+p, func(w http.ResponseWriter, r *http.Request) (any, error) {
			uid := r.PathValue("user_id")
			if uid == "" {
				uid = store.String(user(r), "id")
			}
			if r.URL.Path == "/api/instances" {
				if e := s.permission(r, "Instances"); e != nil {
					return nil, e
				}
				uid = ""
			} else if uid != store.String(user(r), "id") {
				if e := s.permission(r, "Instances"); e != nil {
					return nil, e
				}
			}
			extra := ""
			args := []any{}
			if uid != "" {
				extra = "user_id=$1"
				args = append(args, uid)
			}
			rows, total, e := s.list(r, "instances", extra, args)
			if e != nil {
				return nil, e
			}
			for i, v := range rows {
				rows[i] = s.instanceResponse(v)
			}
			return store.Row{"total": total, "instances": rows}, nil
		})
	}
	s.route("GET /api/instances/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.ownInstance(r, r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		return s.instanceResponse(v), nil
	})
	s.route("PUT /api/instances/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		old, e := s.ownInstance(r, r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		allowed := store.Row{}
		for _, key := range []string{"name", "cpu_usage", "memory_usage", "replicas", "status"} {
			if value, ok := v[key]; ok {
				allowed[key] = value
			}
		}
		if allowed["status"] != nil || allowed["replicas"] != nil {
			if old["status"] == "pending" {
				return nil, fail(409, "实例仍在创建")
			}
			if s.Cluster == nil {
				return nil, fail(503, "Kubernetes 未配置")
			}
			replicas := int32(store.Number(old, "replicas"))
			if n, ok := allowed["replicas"]; ok {
				value := store.Number(store.Row{"n": n}, "n")
				if value < 0 || value > 16 || value != float64(int(value)) {
					return nil, fail(422, "replicas 必须为 0-16")
				}
				replicas = int32(value)
			}
			if status, ok := allowed["status"]; ok {
				if status == "running" {
					replicas = 1
				} else if status == "stopped" {
					replicas = 0
				} else {
					return nil, fail(422, "status 只能为 running/stopped")
				}
			}
			if e = s.Cluster.Scale(r.Context(), store.String(old, "namespace"), store.String(old, "id"), replicas); e != nil {
				return nil, e
			}
			allowed["replicas"] = replicas
			allowed["status"] = "running"
			if replicas == 0 {
				allowed["status"] = "stopped"
			}
		}
		out, e := store.Write(r.Context(), s.DB, "instances", store.String(old, "id"), allowed)
		if e != nil {
			return nil, e
		}
		return s.instanceResponse(out), nil
	})
	for _, action := range []string{"start", "stop"} {
		s.route("POST /api/instances/{id}/"+action, func(w http.ResponseWriter, r *http.Request) (any, error) {
			v, e := s.ownInstance(r, r.PathValue("id"))
			if e != nil {
				return nil, e
			}
			if s.Cluster == nil {
				return nil, fail(503, "Kubernetes 未配置")
			}
			if store.String(v, "status") == "pending" {
				return nil, fail(409, "实例仍在创建")
			}
			replicas, status := int32(1), "running"
			if action == "stop" {
				replicas, status = 0, "stopped"
			}
			if e = s.Cluster.Scale(r.Context(), store.String(v, "namespace"), store.String(v, "id"), replicas); e != nil {
				return nil, e
			}
			v, e = store.Write(r.Context(), s.DB, "instances", store.String(v, "id"), store.Row{"status": status, "replicas": replicas})
			if e != nil {
				return nil, e
			}
			return s.instanceResponse(v), nil
		})
	}
	s.route("DELETE /api/instances/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.ownInstance(r, r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		if store.String(v, "status") == "pending" {
			return nil, fail(409, "实例仍在创建")
		}
		if s.Cluster == nil {
			return nil, fail(503, "Kubernetes 未配置")
		}
		if e = s.Cluster.Delete(r.Context(), store.String(v, "namespace"), store.String(v, "id")); e != nil {
			return nil, e
		}
		return store.Row{}, store.Delete(r.Context(), s.DB, "instances", store.String(v, "id"))
	})
	s.route("POST /api/instances/delete/all", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := admin(r); e != nil {
			return nil, e
		}
		if s.Cluster == nil {
			return nil, fail(503, "Kubernetes 未配置")
		}
		rows, e := store.Query(r.Context(), s.DB, "SELECT * FROM instances")
		if e != nil {
			return nil, e
		}
		deleted := 0
		errs := []store.Row{}
		for _, v := range rows {
			if store.String(v, "status") == "pending" {
				errs = append(errs, store.Row{"id": v["id"], "error": "实例仍在创建"})
				continue
			}
			e = s.Cluster.Delete(r.Context(), store.String(v, "namespace"), store.String(v, "id"))
			if e == nil {
				e = store.Delete(r.Context(), s.DB, "instances", store.String(v, "id"))
			}
			if e != nil {
				errs = append(errs, store.Row{"id": v["id"], "error": e.Error()})
			} else {
				deleted++
			}
		}
		return store.Row{"total": len(rows), "deleted": deleted, "failed": len(errs), "errors": errs}, nil
	})
	s.route("POST /api/instances/start/batch-template", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := admin(r); e != nil {
			return nil, e
		}
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		tmpl, e := store.Get(r.Context(), s.DB, "templates", store.String(v, "template_id"))
		if e != nil {
			return nil, e
		}
		ids := store.Strings(v["user_ids"])
		if len(ids) == 0 || len(ids) > 500 {
			return nil, fail(422, "user_ids 数量必须为 1-500")
		}
		base := store.Object(tmpl["config"])
		if e = normalizeTemplate(base); e != nil {
			return nil, e
		}
		if e = require(base, "repo_id", "repo_tag"); e != nil {
			return nil, e
		}
		if e = cluster.Validate(base); e != nil {
			return nil, fail(422, e.Error())
		}
		for _, id := range ids {
			if _, e = store.Get(r.Context(), s.DB, "users", id); e != nil {
				return nil, e
			}
		}
		tasks := []store.Row{}
		seen := map[string]bool{}
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			copy := store.Row{}
			for k, x := range base {
				copy[k] = x
			}
			copy["user_id"], copy["name"] = id, tmpl["name"]
			out, e := s.enqueue(r, copy)
			if e != nil {
				return nil, e
			}
			tasks = append(tasks, out)
		}
		return store.Row{"template_id": tmpl["id"], "queued_count": len(tasks), "tasks": tasks}, nil
	})
	s.route("GET /api/instances/queue/task/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if s.Redis != nil {
			if raw, e := s.Redis.Get(r.Context(), "instance_task:"+r.PathValue("id")).Bytes(); e == nil {
				var cached store.Row
				if json.Unmarshal(raw, &cached) == nil && cached["user_id"] == user(r)["id"] && (cached["status"] == "completed" || cached["status"] == "failed") {
					return taskResponse(cached), nil
				}
			}
		}
		v, e := store.Get(r.Context(), s.DB, "instance_tasks", r.PathValue("id"))
		if e != nil {
			return nil, e
		}
		if v["user_id"] != user(r)["id"] {
			return nil, fail(403, "无权查看任务")
		}
		return taskResponse(v), nil
	})
	s.route("GET /api/instances/queue/tasks", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rs, e := store.Query(r.Context(), s.DB, "SELECT * FROM instance_tasks WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1000", user(r)["id"])
		if e != nil {
			return nil, e
		}
		for i, v := range rs {
			rs[i] = taskResponse(v)
		}
		return store.Row{"total": len(rs), "tasks": rs}, nil
	})
	s.route("GET /api/instances/queue/status", func(w http.ResponseWriter, r *http.Request) (any, error) {
		rs, e := store.Query(r.Context(), s.DB, "SELECT status,count(*) AS n FROM instance_tasks WHERE user_id=$1 GROUP BY status", user(r)["id"])
		out := store.Row{"pending": 0, "processing": 0, "completed": 0, "failed": 0}
		total := 0
		for _, v := range rs {
			out[store.String(v, "status")] = v["n"]
			total += int(store.Number(v, "n"))
		}
		out["total"] = total
		return out, e
	})
}

func normalizeTemplate(v store.Row) error {
	for _, key := range []string{"cpu_limit", "storage_limit", "gpu_limit", "gpu_memory_limit"} {
		if text, ok := v[key].(string); ok {
			n, e := strconv.ParseFloat(text, 64)
			if e != nil {
				return fail(422, key+" 无效")
			}
			v[key] = n
		}
	}
	if v["memory_limit"] == nil {
		v["memory_limit"] = v["memory_limit_str"]
	}
	if value, ok := v["memory_limit"]; ok && value != nil {
		if _, text := value.(string); !text {
			v["memory_limit"] = fmt.Sprintf("%dMi", int(store.Number(v, "memory_limit")))
		}
	}
	if value, ok := v["storage"]; ok {
		if text, isText := value.(string); isText {
			n, e := strconv.ParseFloat(text, 64)
			if e == nil {
				v["storage_limit"] = n
			} else {
				q, e := resource.ParseQuantity(text)
				if e != nil {
					return fail(422, "storage 无效")
				}
				v["storage_limit"] = float64(q.Value()) / float64(1<<30)
			}
		} else {
			v["storage_limit"] = value
		}
	}
	return nil
}
