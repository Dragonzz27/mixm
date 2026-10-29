package service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Bindings struct{ *Core }

// contentTree 将分类与内容整理为前端可用的树形结构。
func contentTree(rows []model.Record, parent string, depth, max int) []model.Record {
	out := []model.Record{}
	if depth > max {
		return out
	}

	for _, row := range rows {
		if model.String(row, "parent_id") == parent {
			row["depth"], row["type"] = depth, "category"
			row["children"] = contentTree(rows, model.String(row, "id"), depth+1, max)
			out = append(out, row)
		}
	}

	return out
}

// Create 创建内容绑定并返回结果。
func (s *Bindings) Create(ctx context.Context, actor model.Actor, input dto.BindingInput) (any, error) {
	if err := s.permission(ctx, actor, "Bindings"); err != nil {
		return nil, err
	}

	v := input.Fields()
	var err error
	if err = require(v, "content_id", "category_id"); err != nil {
		return nil, err
	}

	return repository.Write(ctx, s.DB, "content_categories", "", v)
}

// List 按权限与查询条件返回内容绑定列表。
func (s *Bindings) List(ctx context.Context, query url.Values) (any, error) {
	rows, total, err := repository.ListResources(ctx, s.DB, model.Actor{}, query, "content_categories", "", nil)
	return model.Record{"items": rows, "total": total}, err
}

// Detail 读取指定内容绑定的详情。
func (s *Bindings) Detail(ctx context.Context, id string) (any, error) {
	return repository.Get(ctx, s.DB, "content_categories", id)
}

// Delete 删除指定内容绑定及其必要关联。
func (s *Bindings) Delete(ctx context.Context, actor model.Actor, id string) (any, error) {
	if err := s.permission(ctx, actor, "Bindings"); err != nil {
		return nil, err
	}

	return model.Record{}, repository.Delete(ctx, s.DB, "content_categories", id)
}

// DeleteForContent 清除指定内容的分类绑定。
func (s *Bindings) DeleteForContent(ctx context.Context, actor model.Actor, id string) (any, error) {
	if err := s.permission(ctx, actor, "Bindings"); err != nil {
		return nil, err
	}

	_, err := repository.ClearContentBindings(ctx, s.DB, id)

	return model.Record{}, err
}

// Batch 按分类批量绑定或解除内容关联。
func (s *Bindings) Batch(
	ctx context.Context,
	actor model.Actor,
	input dto.BindingInput,
	categoryID string,
	unbind bool,
) (any, error) {
	if err := s.permission(ctx, actor, "Bindings"); err != nil {
		return nil, err
	}

	v := input.Fields()
	var err error
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	ids := model.Strings(v["content_ids"])

	for _, id := range ids {
		if !unbind {
			_, err = repository.BindCategoryContent(ctx, tx, uuid.NewString(), id, categoryID)
		} else {
			_, err = repository.UnbindCategoryContent(ctx, tx, categoryID, id)
		}

		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	out := model.Record{"count": len(ids), "message": "操作成功"}

	return out, nil
}

// Counts 统计分类中的内容绑定数量。
func (s *Bindings) Counts(ctx context.Context, query url.Values) (any, error) {
	var err error
	out := model.Record{}
	ids := strings.Split(query.Get("category_ids"), ",")

	for _, id := range ids {
		out[id] = 0
	}

	rs, err := repository.BindingCounts(ctx, s.DB, ids)

	for _, v := range rs {
		out[model.String(v, "category_id")] = v["n"]
	}

	return out, err
}

// CategoryContents 读取指定分类关联的内容列表。
func (s *Bindings) CategoryContents(ctx context.Context, actor model.Actor, query url.Values, id string) (any, error) {
	cat, err := repository.Get(ctx, s.DB, "categories", id)
	if err != nil {
		return nil, err
	}

	rows, total, err := repository.CategoryContents(ctx, s.DB, actor, query, id)

	return model.Record{
		"category": cat,
		"total":    total,
		"contents": rows,
	}, err
}

// ContentTree 返回绑定内容与分类的树形结果。
func (s *Bindings) ContentTree(ctx context.Context, query url.Values, id string) (any, error) {
	maxDepth := 3
	if text := query.Get("max_depth"); text != "" {
		n, err := strconv.Atoi(text)
		if err != nil || n < 1 || n > 20 {
			return nil, fail(http.StatusUnprocessableEntity, "max_depth 必须为 1-20")
		}

		maxDepth = n
	}

	rs, err := repository.ActiveCategories(ctx, s.DB)
	if err != nil {
		return nil, err
	}

	for _, row := range rs {
		if row["id"] == id {
			row["children"] = contentTree(rs, id, 2, maxDepth)
			row["type"] = "category"
			row["depth"] = 1

			return row, nil
		}
	}

	return nil, fail(http.StatusNotFound, "分类不存在")
}
