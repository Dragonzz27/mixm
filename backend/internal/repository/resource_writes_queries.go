package repository

import (
	"context"

	"mixm/backend/internal/model"
)

// CategoryCycle 检查分类父子关系是否会形成循环。
func CategoryCycle(ctx context.Context, db *DB, table, parentID, id string) (model.Record, error) {
	return One(
		ctx,
		db,
		"WITH RECURSIVE c AS (SELECT id,parent_id FROM "+table+" WHERE id=$1 UNION SELECT x.id,x.parent_id FROM "+table+" x JOIN c ON x.id=c.parent_id) SELECT EXISTS(SELECT 1 FROM c WHERE id=$2) AS cycle",
		parentID,
		id,
	)
}

// ClearResourceLinks 清除资源关联表中的指定记录。
func ClearResourceLinks(ctx context.Context, db *DB, prefix, rt, id string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM "+rt+" WHERE "+prefix+"_id=$1", id)
}

// InsertResourceLink 写入一条资源关联关系。
func InsertResourceLink(ctx context.Context, db *DB, col, prefix, rt, id, targetID string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO "+rt+" ("+prefix+"_id,"+col+") VALUES($1,$2)", id, targetID)
}

// ClearQuestionChildren 删除题目的指定子资源。
func ClearQuestionChildren(ctx context.Context, db *DB, rt, questionID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM "+rt+" WHERE question_id=$1", questionID)
}

// ClearPracticeQuestions 清空练习与题目的关联。
func ClearPracticeQuestions(ctx context.Context, db *DB, practiceID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM practice_questions WHERE practice_id=$1", practiceID)
}
