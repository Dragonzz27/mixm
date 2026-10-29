package repository

import (
	"context"
	"fmt"
	"strings"

	"mixm/backend/internal/model"
)

// QuestionCandidates 按筛选条件查询可用于组卷的候选题目。
func QuestionCandidates(ctx context.Context, db *DB, actor model.Actor, v model.Record) ([]model.Record, error) {
	condition, args, err := Visibility(ctx, db, actor, "questions", 1)
	if err != nil {
		return nil, err
	}

	where := []string{"review_status='approved'", condition}
	add := func(expr string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(expr, len(args)))
	}
	if id := model.String(v, "course_id"); id != "" {
		add(
			"EXISTS(SELECT 1 FROM question_tags t WHERE t.question_id=questions.id AND t.category_id IN (WITH RECURSIVE c AS (SELECT id FROM categories WHERE id=$%d UNION SELECT x.id FROM categories x JOIN c ON x.parent_id=c.id) SELECT id FROM c))",
			id,
		)
	}

	if tags := model.Strings(v["knowledge_tag_ids"]); len(tags) > 0 {
		add("EXISTS(SELECT 1 FROM question_tags t WHERE t.question_id=questions.id AND t.category_id=ANY($%d))", tags)
	}

	if v["difficulty"] != nil {
		add("difficulty=$%d", int(model.Number(v, "difficulty")))
	}

	types := model.Strings(v["question_types"])

	for i, t := range types {
		types[i] = model.BankType(t)
	}

	if len(types) > 0 {
		add("type=ANY($%d)", types)
	}

	order := "random()"
	if id := model.String(v, "practice_id"); id != "" {
		if _, err := ResourceDetail(ctx, db, actor, "practices", id); err != nil {
			return nil, err
		}

		add("id IN (SELECT question_id FROM practice_questions WHERE practice_id=$%d)", id)
		args = append(args, id)
		order = fmt.Sprintf("(SELECT sort_order FROM practice_questions WHERE practice_id=$%d AND question_id=questions.id)", len(args))
	}

	rows, err := Query(ctx, db, "SELECT * FROM questions WHERE "+strings.Join(where, " AND ")+" ORDER BY "+order+" LIMIT 1000", args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
