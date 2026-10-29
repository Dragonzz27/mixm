package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Instances struct{ *Core }

// ownInstance 读取实例并校验当前用户的访问权限。
func (s *Instances) ownInstance(ctx context.Context, actor model.Actor, id string) (model.Record, error) {
	v, err := repository.Get(ctx, s.DB, "instances", id)
	if err != nil {
		return nil, err
	}

	if v["user_id"] != actor.ID {
		if err = s.permission(ctx, actor, "Instances"); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// instanceResponse 补充实例访问地址、SSH 凭据和前端展示字段。
func (s *Instances) instanceResponse(v model.Record) model.Record {
	id := model.String(v, "id")
	v["endpoint"] = "/instance/" + id
	cfg := model.Object(v["config"])
	v["ssh_host"], v["ssh_port"] = s.Config.SSHHost, s.Config.SSHPort
	v["ssh_password"] = cfg["password"]
	v["hostIp"] = s.Config.CodeHost
	v["code_server_host"] = s.Config.CodeHost
	v["ssh_command"] = fmt.Sprintf("ssh %s@%s -p %d", model.String(v, "ssh_username"), s.Config.SSHHost, s.Config.SSHPort)
	v["ssh_connection"] = v["ssh_command"]
	v["access_url"] = "http://" + s.Config.CodeHost + "/instance/" + id + "/"
	v["code_server_url"] = v["access_url"]
	v["password"] = cfg["password"]

	return v
}

// taskResponse 整理任务状态及其关联的实例配置。
func taskResponse(t model.Record) model.Record {
	v := model.Object(t["payload"])

	for _, k := range []string{
		"user_id",
		"status",
		"instance_id",
		"error",
		"created_at",
		"started_at",
		"completed_at",
	} {
		v[k] = t[k]
	}

	v["task_id"] = t["id"]
	v["instance_name"] = v["name"]

	return v
}

// enqueue 校验资源和镜像权限，事务创建实例及持久化任务。
func (s *Instances) enqueue(ctx context.Context, actor model.Actor, v model.Record) (model.Record, error) {
	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	// 确定实例所属用户，跨用户创建时检查管理权限。
	uid := model.String(v, "user_id")
	if uid == "" {
		uid = actor.ID
	}

	if uid != actor.ID {
		if err := s.permission(ctx, actor, "Instances"); err != nil {
			return nil, err
		}
	}

	u, err := repository.Get(ctx, s.DB, "users", uid)
	if err != nil {
		return nil, err
	}

	if !model.Bool(u, "is_active") {
		return nil, fail(http.StatusUnprocessableEntity, "用户已停用")
	}

	// 统一资源配置并检查镜像仓库的可用范围。
	v["user_id"] = uid
	if err = cluster.Validate(v); err != nil {
		return nil, fail(http.StatusUnprocessableEntity, err.Error())
	}

	if model.String(v, "repo_id") != "" {
		repo, err := repository.Get(ctx, s.DB, "image_repositories", model.String(v, "repo_id"))
		if err != nil {
			return nil, err
		}

		if !model.Bool(repo, "is_public") && repo["user_id"] != uid && !actor.IsSuperuser {
			return nil, fail(http.StatusForbidden, "无权使用镜像")
		}

		if err = require(v, "repo_tag"); err != nil {
			return nil, err
		}

		v["image"] = strings.TrimSuffix(s.Config.RegistryPrefix, "/") + "/" + model.String(repo, "name") + ":" + model.String(v, "repo_tag")
	}

	// 生成实例名称和访问凭据，整理镜像及工作目录信息。
	if model.String(v, "image") == "" {
		v["image"] = "ubuntu:24.04"
	}

	if model.String(v, "name") == "" {
		v["name"] = "instance"
	}

	v["name"] = fmt.Sprintf("%s-%d", v["name"], time.Now().UnixNano())
	delete(v, "ssh_username")
	v["ssh_port"] = s.Config.SSHPort
	v["status"] = "pending"
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return nil, err
	}

	v["config"] = model.JSON(model.Record{
		"password":    hex.EncodeToString(b),
		"working_dir": "/workspace",
		"image":       v["image"],
	})
	memory, _ := resource.ParseQuantity(model.String(v, "memory_limit"))
	v["memory_limit_mb"] = memory.Value() / (1 << 20)

	// 实例记录和创建任务必须在同一事务中保存。
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	instance, err := repository.Write(ctx, tx, "instances", "", v)
	if err != nil {
		return nil, err
	}

	// 使用实际保存的实例 ID 生成 SSH 标识，保持访问入口一致。
	id := model.String(instance, "id")
	v["ssh_username"] = model.String(u, "username") + strings.ReplaceAll(id, "-", "")[:8]
	if _, err = repository.Write(ctx, tx, "instances", id, model.Record{"ssh_username": v["ssh_username"]}); err != nil {
		return nil, err
	}

	// 持久化任务后统一提交，由发布器可靠投递到消息队列。
	task, err := repository.Write(
		ctx,
		tx,
		"instance_tasks",
		"",
		model.Record{
			"user_id":     uid,
			"instance_id": instance["id"],
			"payload":     v,
		},
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return model.Record{
		"task_id":       task["id"],
		"instance_id":   instance["id"],
		"instance_name": instance["name"],
		"user_id":       uid,
		"status":        "pending",
		"message":       "实例已加入创建队列",
		"config": model.Record{
			"repo_id":          v["repo_id"],
			"repo_tag":         v["repo_tag"],
			"cpu_limit":        v["cpu_limit"],
			"memory_limit":     v["memory_limit"],
			"storage_limit":    v["storage_limit"],
			"gpu_limit":        v["gpu_limit"],
			"gpu_memory_limit": v["gpu_memory_limit"],
		},
	}, nil
}

// normalizeTemplate 统一模板中的资源字段类型与容量单位。
func normalizeTemplate(v model.Record) error {
	for _, key := range []string{"cpu_limit", "storage_limit", "gpu_limit", "gpu_memory_limit"} {
		if text, ok := v[key].(string); ok {
			n, err := strconv.ParseFloat(text, 64)
			if err != nil {
				return fail(http.StatusUnprocessableEntity, key+" 无效")
			}

			v[key] = n
		}
	}

	if v["memory_limit"] == nil {
		v["memory_limit"] = v["memory_limit_str"]
	}

	if value, ok := v["memory_limit"]; ok && value != nil {
		if _, text := value.(string); !text {
			v["memory_limit"] = fmt.Sprintf("%dMi", int(model.Number(v, "memory_limit")))
		}
	}

	if value, ok := v["storage"]; ok {
		if text, isText := value.(string); isText {
			n, err := strconv.ParseFloat(text, 64)
			if err == nil {
				v["storage_limit"] = n
			} else {
				q, err := resource.ParseQuantity(text)
				if err != nil {
					return fail(http.StatusUnprocessableEntity, "storage 无效")
				}

				v["storage_limit"] = float64(q.Value()) / float64(1<<30)
			}
		} else {
			v["storage_limit"] = value
		}
	}

	return nil
}

// Create 创建实例并返回结果。
func (s *Instances) Create(ctx context.Context, actor model.Actor, input dto.InstanceInput) (any, error) {
	v := input.Fields()
	var err error
	out, err := s.enqueue(ctx, actor, v)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// List 按权限与查询条件返回实例列表。
func (s *Instances) List(
	ctx context.Context,
	actor model.Actor,
	query url.Values,
	userID string,
	all bool,
) (any, error) {
	uid := userID
	if uid == "" {
		uid = actor.ID
	}

	if all {
		if err := s.permission(ctx, actor, "Instances"); err != nil {
			return nil, err
		}

		uid = ""
	} else if uid != actor.ID {
		if err := s.permission(ctx, actor, "Instances"); err != nil {
			return nil, err
		}
	}

	extra := ""
	args := []any{}
	if uid != "" {
		extra = "user_id=$1"
		args = append(args, uid)
	}

	rows, total, err := s.list(ctx, actor, query, "instances", extra, args)
	if err != nil {
		return nil, err
	}

	for i, v := range rows {
		rows[i] = s.instanceResponse(v)
	}

	return model.Record{"total": total, "instances": rows}, nil
}

// Detail 读取指定实例的详情。
func (s *Instances) Detail(ctx context.Context, actor model.Actor, id string) (any, error) {
	v, err := s.ownInstance(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	return s.instanceResponse(v), nil
}

// Update 校验并更新指定实例。
func (s *Instances) Update(ctx context.Context, actor model.Actor, input dto.InstanceInput, id string) (any, error) {
	old, err := s.ownInstance(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	v := input.Fields()
	allowed := model.Record{}

	for _, key := range []string{"name", "cpu_usage", "memory_usage", "replicas", "status"} {
		if value, ok := v[key]; ok {
			allowed[key] = value
		}
	}

	if allowed["status"] != nil || allowed["replicas"] != nil {
		if old["status"] == "pending" {
			return nil, fail(http.StatusConflict, "实例仍在创建")
		}

		if s.Cluster == nil {
			return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
		}

		replicas := int32(model.Number(old, "replicas"))
		if n, ok := allowed["replicas"]; ok {
			value := model.Number(model.Record{"n": n}, "n")
			if value < 0 || value > 16 || value != float64(int(value)) {
				return nil, fail(http.StatusUnprocessableEntity, "replicas 必须为 0-16")
			}

			replicas = int32(value)
		}

		if status, ok := allowed["status"]; ok {
			if status == "running" {
				replicas = 1
			} else if status == "stopped" {
				replicas = 0
			} else {
				return nil, fail(http.StatusUnprocessableEntity, "status 只能为 running/stopped")
			}
		}

		if err = s.Cluster.Scale(ctx, model.String(old, "namespace"), model.String(old, "id"), replicas); err != nil {
			return nil, err
		}

		allowed["replicas"] = replicas
		allowed["status"] = "running"
		if replicas == 0 {
			allowed["status"] = "stopped"
		}
	}

	out, err := repository.Write(ctx, s.DB, "instances", model.String(old, "id"), allowed)
	if err != nil {
		return nil, err
	}

	return s.instanceResponse(out), nil
}

// Delete 删除指定实例及其必要关联。
func (s *Instances) Delete(ctx context.Context, actor model.Actor, id string) (any, error) {
	v, err := s.ownInstance(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	if model.String(v, "status") == "pending" {
		return nil, fail(http.StatusConflict, "实例仍在创建")
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	if err = s.Cluster.Delete(ctx, model.String(v, "namespace"), model.String(v, "id")); err != nil {
		return nil, err
	}

	return model.Record{}, repository.Delete(ctx, s.DB, "instances", model.String(v, "id"))
}

// DeleteAll 删除当前管理范围内的全部实例。
func (s *Instances) DeleteAll(ctx context.Context, actor model.Actor) (any, error) {
	if err := admin(actor); err != nil {
		return nil, err
	}

	if s.Cluster == nil {
		return nil, fail(http.StatusServiceUnavailable, "Kubernetes 未配置")
	}

	rows, err := repository.AllInstances(ctx, s.DB)
	if err != nil {
		return nil, err
	}

	deleted := 0
	errs := []model.Record{}

	for _, v := range rows {
		if model.String(v, "status") == "pending" {
			errs = append(errs, model.Record{"id": v["id"], "error": "实例仍在创建"})
			continue
		}

		err = s.Cluster.Delete(ctx, model.String(v, "namespace"), model.String(v, "id"))
		if err == nil {
			err = repository.Delete(ctx, s.DB, "instances", model.String(v, "id"))
		}

		if err != nil {
			errs = append(errs, model.Record{"id": v["id"], "error": err.Error()})
		} else {
			deleted++
		}
	}

	return model.Record{
		"total":   len(rows),
		"deleted": deleted,
		"failed":  len(errs),
		"errors":  errs,
	}, nil
}

// Batch 根据模板为选定用户批量创建实例任务。
func (s *Instances) Batch(
	ctx context.Context,
	actor model.Actor,
	input dto.InstanceInput,
	templateID string,
) (any, error) {
	if err := admin(actor); err != nil {
		return nil, err
	}

	v := input.Fields()
	var err error
	tmpl, err := repository.Get(ctx, s.DB, "templates", templateID)
	if err != nil {
		return nil, err
	}

	ids := model.Strings(v["user_ids"])
	if len(ids) == 0 || len(ids) > 500 {
		return nil, fail(http.StatusUnprocessableEntity, "user_ids 数量必须为 1-500")
	}

	base := model.Object(tmpl["config"])
	if err = normalizeTemplate(base); err != nil {
		return nil, err
	}

	if err = require(base, "repo_id", "repo_tag"); err != nil {
		return nil, err
	}

	if err = cluster.Validate(base); err != nil {
		return nil, fail(http.StatusUnprocessableEntity, err.Error())
	}

	for _, id := range ids {
		if _, err = repository.Get(ctx, s.DB, "users", id); err != nil {
			return nil, err
		}
	}

	tasks := []model.Record{}
	seen := map[string]bool{}

	for _, id := range ids {
		if seen[id] {
			continue
		}

		seen[id] = true
		copy := model.Record{}

		for k, x := range base {
			copy[k] = x
		}

		copy["user_id"], copy["name"] = id, tmpl["name"]
		out, err := s.enqueue(ctx, actor, copy)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, out)
	}

	return model.Record{
		"template_id":  tmpl["id"],
		"queued_count": len(tasks),
		"tasks":        tasks,
	}, nil
}

// Task 读取单个实例创建任务的进度。
func (s *Instances) Task(ctx context.Context, actor model.Actor, id string) (any, error) {
	if s.Redis != nil {
		if raw, err := s.Redis.Get(ctx, "instance_task:"+id).Bytes(); err == nil {
			var cached model.Record
			if json.Unmarshal(raw, &cached) == nil && cached["user_id"] == actor.ID && (cached["status"] == "completed" || cached["status"] == "failed") {
				return taskResponse(cached), nil
			}
		}
	}

	v, err := repository.Get(ctx, s.DB, "instance_tasks", id)
	if err != nil {
		return nil, err
	}

	if v["user_id"] != actor.ID {
		return nil, fail(http.StatusForbidden, "无权查看任务")
	}

	return taskResponse(v), nil
}

// Tasks 读取用户可见的实例创建任务列表。
func (s *Instances) Tasks(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	rs, total, err := s.list(ctx, actor, query, "instance_tasks", "user_id=$1", []any{actor.ID})
	if err != nil {
		return nil, err
	}

	for i, v := range rs {
		rs[i] = taskResponse(v)
	}

	return model.Record{"total": total, "tasks": rs}, nil
}

// Summary 汇总实例创建任务的状态数量。
func (s *Instances) Summary(ctx context.Context, actor model.Actor) (any, error) {
	rs, err := repository.TaskSummary(ctx, s.DB, actor.ID)
	out := model.Record{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}
	total := 0

	for _, v := range rs {
		out[model.String(v, "status")] = v["n"]
		total += int(model.Number(v, "n"))
	}

	out["total"] = total

	return out, err
}
