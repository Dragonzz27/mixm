package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Assignments struct{ *Core }

// submissionResponse 整理作业提交的文件信息和下载地址。
func submissionResponse(v model.Record) model.Record {
	if p := model.String(v, "file_path"); p != "" {
		v["file_url"] = "/api/assignments/files/" + filepath.Base(p)
	}

	return v
}

// Mine 读取当前用户的作业列表。
func (s *Assignments) Mine(ctx context.Context, actor model.Actor, query url.Values) (any, error) {
	extra := "user_id=$1"
	args := []any{actor.ID}
	if id := query.Get("assignment_id"); id != "" {
		extra += " AND assignment_id=$2"
		args = append(args, id)
	}

	rs, total, err := s.list(ctx, actor, query, "submissions", extra, args)

	for i, v := range rs {
		rs[i] = submissionResponse(v)
	}

	return model.Record{"items": rs, "total": total}, err
}

// Submit 校验作业与附件信息，保存当前用户的提交记录。
func (s *Assignments) Submit(ctx context.Context, actor model.Actor, input dto.SubmissionInput, id string) (any, error) {
	v := input.Fields()
	v["assignment_id"] = id
	var err error
	a, err := repository.ResourceDetail(ctx, s.DB, actor, "assignments", id)
	if err != nil {
		return nil, err
	}

	deadline, _ := a["deadline"].(time.Time)
	if model.String(a, "status") != "active" || deadline.Before(time.Now()) {
		return nil, fail(http.StatusBadRequest, "作业已截止")
	}

	v["user_id"], v["status"] = actor.ID, "submitted"
	delete(v, "grade")
	delete(v, "feedback")
	delete(v, "reviewed_at")
	if p := model.String(v, "file_path"); p != "" {
		if filepath.Clean(p) != filepath.Join(s.Config.UploadDir, filepath.Base(p)) {
			return nil, fail(http.StatusUnprocessableEntity, "文件路径无效")
		}

		if _, err = os.Stat(p); err != nil {
			return nil, fail(http.StatusUnprocessableEntity, "提交文件不存在")
		}
	}

	if v["content"] == nil && v["file_path"] == nil {
		return nil, fail(http.StatusUnprocessableEntity, "需要提交内容或文件")
	}

	row, err := repository.Write(ctx, s.DB, "submissions", "", v)
	if err != nil {
		return nil, err
	}

	return submissionResponse(row), nil
}

// Submissions 按权限查询作业提交记录。
func (s *Assignments) Submissions(ctx context.Context, actor model.Actor, query url.Values, id string) (any, error) {
	if err := s.permission(ctx, actor, "Assignments"); err != nil {
		return nil, err
	}

	rs, total, err := s.list(ctx, actor, query, "submissions", "assignment_id=$1", []any{id})

	for i, v := range rs {
		rs[i] = submissionResponse(v)
	}

	return model.Record{"total": total, "submissions": rs}, err
}

// Review 校验批改权限并保存成绩与反馈。
func (s *Assignments) Review(ctx context.Context, actor model.Actor, input dto.SubmissionInput, id string) (any, error) {
	if err := s.permission(ctx, actor, "Assignments"); err != nil {
		return nil, err
	}

	v := input.Fields()
	var err error
	allowed := model.Record{"reviewed_at": time.Now(), "status": "reviewed"}

	for _, k := range []string{"grade", "feedback", "status"} {
		if x, ok := v[k]; ok {
			allowed[k] = x
		}
	}

	out, err := repository.Write(ctx, s.DB, "submissions", id, allowed)
	if err != nil {
		return nil, err
	}

	return submissionResponse(out), nil
}

// DeleteSubmission 删除指定作业提交记录。
func (s *Assignments) DeleteSubmission(ctx context.Context, actor model.Actor, id string) (any, error) {
	if err := s.permission(ctx, actor, "Assignments"); err != nil {
		return nil, err
	}

	return model.Record{}, repository.Delete(ctx, s.DB, "submissions", id)
}

// Export 生成作业提交记录的导出文件。
func (s *Assignments) Export(ctx context.Context, actor model.Actor, id string) (any, error) {
	if err := s.permission(ctx, actor, "Assignments"); err != nil {
		return nil, err
	}

	rows, err := repository.SubmissionExport(ctx, s.DB, id)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(s.Config.ExportDir, 0755); err != nil {
		return nil, err
	}

	name := "grades-" + uuid.NewString() + ".csv"
	f, err := os.Create(filepath.Join(s.Config.ExportDir, name))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	writer := csv.NewWriter(f)
	keys := []string{
		"id",
		"user_id",
		"username",
		"file_name",
		"submitted_at",
		"status",
		"grade",
		"feedback",
	}
	if err = writer.Write(keys); err != nil {
		return nil, err
	}

	for _, v := range rows {
		line := []string{}

		for _, k := range keys {
			text := ""
			if v[k] != nil {
				text = fmt.Sprint(v[k])
			}

			if len(text) > 0 && (text[0] == '=' || text[0] == '+' || text[0] == '-' || text[0] == '@') {
				text = "'" + text
			}

			line = append(line, text)
		}

		if err = writer.Write(line); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		return nil, err
	}

	return model.Record{"file_url": "/static/export/" + name, "filename": name}, nil
}
