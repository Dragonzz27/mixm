package repository

import (
	"context"

	"mixm/backend/internal/model"
)

func CategoryCycle(ctx context.Context, db *DB, table, parentID, id string) (model.Record, error) {
	return One(ctx, db, "WITH RECURSIVE c AS (SELECT id,parent_id FROM "+table+" WHERE id=$1 UNION SELECT x.id,x.parent_id FROM "+table+" x JOIN c ON x.id=c.parent_id) SELECT EXISTS(SELECT 1 FROM c WHERE id=$2) AS cycle", parentID, id)
}
func ClearResourceLinks(ctx context.Context, db *DB, prefix, rt, id string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM "+rt+" WHERE "+prefix+"_id=$1", id)
}
func InsertResourceLink(ctx context.Context, db *DB, col, prefix, rt, id, targetID string) (Result, error) {
	return db.Exec(ctx, "INSERT INTO "+rt+" ("+prefix+"_id,"+col+") VALUES($1,$2)", id, targetID)
}
func ClearQuestionChildren(ctx context.Context, db *DB, rt, questionID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM "+rt+" WHERE question_id=$1", questionID)
}
func ClearPracticeQuestions(ctx context.Context, db *DB, practiceID string) (Result, error) {
	return db.Exec(ctx, "DELETE FROM practice_questions WHERE practice_id=$1", practiceID)
}
