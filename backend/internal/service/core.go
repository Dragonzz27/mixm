package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/client"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Core struct {
	Config  config.Config
	DB      *repository.DB
	Redis   *redis.Client
	Cluster *cluster.Client
	Clients *client.Clients
}

// NewCore 组装配置、数据库、缓存、集群与外部接口依赖。
func NewCore(c config.Config, db *repository.DB, cache *redis.Client, k *cluster.Client) *Core {
	return &Core{
		Config:  c,
		DB:      db,
		Redis:   cache,
		Cluster: k,
		Clients: client.New(c),
	}
}

// fail 构造携带 HTTP 状态和公开消息的业务错误。
func fail(status int, message string) error {
	return &apperror.Error{Status: status, Detail: message}
}

// admin 校验当前用户是否为超级管理员。
func admin(actor model.Actor) error {
	if !actor.IsSuperuser {
		return fail(http.StatusForbidden, "需要管理员权限")
	}

	return nil
}

// require 检查必填字符串字段是否为空。
func require(v model.Record, fields ...string) error {
	for _, field := range fields {
		if strings.TrimSpace(model.String(v, field)) == "" {
			return fail(http.StatusUnprocessableEntity, fmt.Sprintf("%s 不能为空", field))
		}
	}

	return nil
}

// permission 检查当前用户是否拥有指定功能权限。
func (s *Core) permission(ctx context.Context, actor model.Actor, key string) error {
	allowed, err := repository.HasPermission(ctx, s.DB, actor, key)
	if err != nil {
		return err
	}

	if !allowed {
		return fail(http.StatusForbidden, "没有操作权限")
	}

	return nil
}

// list 按资源类型、用户权限和分页条件查询列表。
func (s *Core) list(
	ctx context.Context,
	actor model.Actor,
	query url.Values,
	table, extra string,
	args []any,
) ([]model.Record, int, error) {
	return repository.ListResources(ctx, s.DB, actor, query, table, extra, args)
}

// enrich 补充资源关联信息与展示字段。
func (s *Core) enrich(ctx context.Context, db *repository.DB, table string, v model.Record) (model.Record, error) {
	return repository.Enrich(ctx, db, table, v)
}
