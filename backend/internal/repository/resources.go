package repository

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
)

// HasPermission 判断用户是否拥有指定路由权限。
func HasPermission(ctx context.Context, db *DB, actor model.Actor, key string) (bool, error) {
	if actor.IsSuperuser {
		return true, nil
	}

	v, err := One(
		ctx,
		db,
		"SELECT EXISTS(SELECT 1 FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=$1 AND jsonb_exists(r.route_keys,$2)) AS allowed",
		actor.ID,
		key,
	)

	return model.Bool(v, "allowed"), err
}

// Visibility 按资源类型构造当前用户的可见范围。
func Visibility(ctx context.Context, db *DB, actor model.Actor, table string, index int) (string, []any, error) {
	if actor.IsSuperuser {
		return "true", nil, nil
	}

	prefix := map[string]string{
		"assignments": "assignment",
		"questions":   "question",
		"practices":   "practice",
	}[table]
	if prefix == "" {
		return "true", nil, nil
	}

	condition := fmt.Sprintf(
		"(NOT EXISTS(SELECT 1 FROM %s_roles vr WHERE vr.%s_id=%s.id) OR EXISTS(SELECT 1 FROM %s_roles vr JOIN user_roles ur ON ur.role_id=vr.role_id WHERE vr.%s_id=%s.id AND ur.user_id=$%d))",
		prefix,
		prefix,
		table,
		prefix,
		prefix,
		table,
		index,
	)
	if table == "questions" {
		allowed, err := HasPermission(ctx, db, actor, "Questions")
		if err != nil {
			return "", nil, err
		}

		if !allowed {
			condition += " AND review_status='approved'"
		}
	}

	return condition, []any{actor.ID}, nil
}

// ResourceDetail 读取资源详情并校验访问权限。
func ResourceDetail(ctx context.Context, db *DB, actor model.Actor, table, id string) (model.Record, error) {
	if Columns[table] == nil {
		return nil, fmt.Errorf("unknown resource")
	}

	condition, args, err := Visibility(ctx, db, actor, table, 2)
	if err != nil {
		return nil, err
	}

	return One(ctx, db, "SELECT * FROM "+table+" WHERE id=$1 AND "+condition, append([]any{id}, args...)...)
}

// parsePage 解析分页参数，返回偏移量和每页数量。
func parsePage(query url.Values) (int, int, error) {
	p, err := dto.ParsePagination(query)
	if err != nil {
		return 0, 0, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: err.Error()}
	}

	return p.Offset, p.Limit, nil
}

// fail 构造携带 HTTP 状态和公开消息的业务错误。
func fail(status int, message string) error {
	return &apperror.Error{Status: status, Detail: message}
}

// ListResources 结合可见范围、筛选和分页条件查询资源。
func ListResources(
	ctx context.Context,
	db *DB,
	actor model.Actor,
	query url.Values,
	table, extra string,
	args []any,
) ([]model.Record, int, error) {
	skip, limit, err := parsePage(query)
	if err != nil {
		return nil, 0, err
	}

	args = append([]any{}, args...)
	condition, visibilityArgs, err := Visibility(ctx, db, actor, table, len(args)+1)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, visibilityArgs...)
	where := []string{condition}
	if extra != "" {
		where = append(where, extra)
	}

	q := query
	if table == "assignments" && q.Has("active") {
		active, err := strconv.ParseBool(q.Get("active"))
		if err != nil {
			return nil, 0, fail(http.StatusUnprocessableEntity, "active 无效")
		}

		if active {
			where = append(where, "status='active' AND deadline>now()")
		}
	}

	if table == "categories" && q.Get("root_only") == "true" {
		where = append(where, "parent_id IS NULL AND is_active=true")
	}

	if table == "contents" {
		if id := q.Get("unbound_category_id"); id != "" {
			args = append(args, id)
			where = append(
				where,
				fmt.Sprintf(
					"NOT EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id=$%d)",
					len(args),
				),
			)
		}

		if ids := q.Get("category_ids"); ids != "" {
			args = append(args, strings.Split(ids, ","))
			where = append(
				where,
				fmt.Sprintf(
					"is_active=true AND status='published' AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id=ANY($%d))",
					len(args),
				),
			)
		}
	}

	for _, k := range []string{
		"content_id",
		"category_id",
		"parent_id",
		"content_type",
		"status",
		"template_id",
		"type",
		"review_status",
		"difficulty",
		"is_active",
		"is_ai_generated",
	} {
		if q.Has(k) {
			if _, ok := Columns[table][k]; ok {
				var filter any = q.Get(k)
				if Columns[table][k] == "integer" {
					n, err := strconv.Atoi(q.Get(k))
					if err != nil {
						return nil, 0, fail(http.StatusUnprocessableEntity, k+" 无效")
					}

					filter = n
				}

				if Columns[table][k] == "boolean" {
					b, err := strconv.ParseBool(q.Get(k))
					if err != nil {
						return nil, 0, fail(http.StatusUnprocessableEntity, k+" 无效")
					}

					filter = b
				}

				args = append(args, filter)
				if k == "parent_id" && q.Get(k) == "" {
					where = append(where, "parent_id IS NULL")
					args = args[:len(args)-1]
				} else {
					where = append(where, fmt.Sprintf("%s=$%d", k, len(args)))
				}
			}
		}
	}

	kw := q.Get("keyword")
	if kw == "" {
		kw = q.Get("search")
	}

	if kw != "" {
		field := "name"
		if _, ok := Columns[table]["title"]; ok {
			field = "title"
		}

		if table == "users" {
			field = "username"
		}

		args = append(args, "%"+kw+"%")
		where = append(where, fmt.Sprintf("%s ILIKE $%d", field, len(args)))
	}

	if q.Get("knowledge_tag_ids") != "" && (table == "questions" || table == "practices") {
		prefix := strings.TrimSuffix(table, "s")
		args = append(args, strings.Split(q.Get("knowledge_tag_ids"), ","))
		where = append(
			where,
			fmt.Sprintf(
				"EXISTS(SELECT 1 FROM %s_tags t WHERE t.%s_id=%s.id AND t.category_id=ANY($%d))",
				prefix,
				prefix,
				table,
				len(args),
			),
		)
	}

	clause := strings.Join(where, " AND ")
	count, err := One(ctx, db, "SELECT count(*) AS total FROM "+table+" WHERE "+clause, args...)
	if err != nil {
		return nil, 0, err
	}

	order := "id"
	if _, ok := Columns[table]["created_at"]; ok {
		order = "created_at DESC,id"
	}

	if table == "categories" {
		order = "sort_order,id"
	}

	if table == "submissions" {
		order = "submitted_at DESC,id"
	}

	args = append(args, limit, skip)
	rows, err := Query(
		ctx,
		db,
		fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d", table, clause, order, len(args)-1, len(args)),
		args...,
	)

	return rows, int(model.Number(count, "total")), err
}

// Enrich 补充资源的关联角色、分类或展示信息。
func Enrich(ctx context.Context, db *DB, table string, v model.Record) (model.Record, error) {
	id := v["id"]
	var err error
	prefix := map[string]string{
		"questions":   "question",
		"practices":   "practice",
		"assignments": "assignment",
	}[table]
	if prefix != "" {
		roles, err := Query(ctx, db, "SELECT role_id FROM "+prefix+"_roles WHERE "+prefix+"_id=$1", id)
		if err != nil {
			return nil, err
		}

		ids := []string{}

		for _, role := range roles {
			ids = append(ids, model.String(role, "role_id"))
		}

		v["visible_roles"] = ids
	}

	if table == "questions" || table == "practices" {
		tags, err := Query(
			ctx,
			db,
			"SELECT c.id,c.name FROM categories c JOIN "+prefix+"_tags t ON t.category_id=c.id WHERE t."+prefix+"_id=$1",
			id,
		)
		if err != nil {
			return nil, err
		}

		ids := []string{}

		for _, t := range tags {
			ids = append(ids, model.String(t, "id"))
		}

		v["knowledge_tags"], v["knowledge_tag_ids"] = tags, ids
	}

	switch table {
	case "questions":
		v["options"], err = Query(ctx, db, "SELECT * FROM question_options WHERE question_id=$1 ORDER BY option_key", id)
		if err != nil {
			return nil, err
		}

		v["attachments"], err = Query(ctx, db, "SELECT * FROM question_attachments WHERE question_id=$1", id)
	case "practices":
		rows, err := Query(
			ctx,
			db,
			"SELECT q.* FROM questions q JOIN practice_questions pq ON pq.question_id=q.id WHERE pq.practice_id=$1 ORDER BY pq.sort_order",
			id,
		)
		if err != nil {
			return nil, err
		}

		for i, q := range rows {
			rows[i], err = Enrich(ctx, db, "questions", q)
			if err != nil {
				return nil, err
			}
		}

		v["questions"], v["question_count"] = rows, len(rows)
	case "contents":
		v["categories"], err = Query(
			ctx,
			db,
			"SELECT c.* FROM categories c JOIN content_categories cc ON cc.category_id=c.id WHERE cc.content_id=$1 ORDER BY cc.sort_order",
			id,
		)
	case "assignments":
		v["attachments"] = model.Objects(v["attachments"])
		cnt, err := One(ctx, db, "SELECT count(*) AS n FROM submissions WHERE assignment_id=$1", id)
		if err != nil {
			return nil, err
		}

		v["submission_count"] = cnt["n"]
	}

	return v, err
}

// CategoryContents 读取指定分类关联的内容列表。
func CategoryContents(ctx context.Context, db *DB, actor model.Actor, query url.Values, id string) ([]model.Record, int, error) {
	return ListResources(
		ctx,
		db,
		actor,
		query,
		"contents",
		"is_active=true AND status='published' AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$1 UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))",
		[]any{id},
	)
}

// SimulationContents 查询可用于模拟测试的内容。
func SimulationContents(ctx context.Context, db *DB, courseID string) ([]model.Record, error) {
	query := "SELECT id,title,content FROM contents WHERE status='published' AND is_active=true"
	args := []any{}
	if courseID != "" {
		args = append(args, courseID)
		query += " AND EXISTS(SELECT 1 FROM content_categories cc WHERE cc.content_id=contents.id AND cc.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$1 UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))"
	}

	return Query(ctx, db, query+" LIMIT 100", args...)
}
