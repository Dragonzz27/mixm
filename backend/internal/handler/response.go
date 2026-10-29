package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

// Failure 将业务错误转换为统一的 HTTP 错误响应。
func Failure(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "服务器错误"
	var application *apperror.Error
	var pg *pgconn.PgError

	switch {
	case errors.As(err, &application):
		status, message = application.Status, application.Detail
	case errors.Is(err, repository.ErrNotFound):
		status, message = http.StatusNotFound, "资源不存在"
	case errors.Is(err, repository.ErrInvalidValue):
		status, message = http.StatusUnprocessableEntity, err.Error()
	case errors.As(err, &pg):
		switch pg.Code {
		case "23505":
			status, message = http.StatusConflict, "数据已存在"
		case "23503", "23514", "23502", "22P02":
			status, message = http.StatusUnprocessableEntity, "数据或关联资源无效"
		}
	}

	if status == 500 {
		slog.Error("request failed", "path", c.Request.URL.Path, "error", err)
	}

	if c.Writer.Written() {
		c.Abort()
		return
	}

	c.AbortWithStatusJSON(status, dto.Envelope{Code: status, Message: message})
}

// Success 使用统一响应结构返回成功结果。
func Success(c *gin.Context, status int, value any) {
	switch v := value.(type) {
	case dto.File:
		if v.Path != "" {
			http.ServeFile(c.Writer, c.Request, v.Path)
			return
		}

		if v.Reader != nil {
			defer v.Reader.Close()

			if c.Request.Method == http.MethodHead {
				c.Header("Content-Type", v.ContentType)
				c.Header("Content-Length", strconv.FormatInt(v.Size, 10))
				c.Status(status)

				return
			}

			c.DataFromReader(status, v.Size, v.ContentType, v.Reader, nil)

			return
		}
	}

	c.JSON(status, dto.Envelope{Message: "成功", Data: value})
}

// actor 从请求上下文获取已认证的用户身份。
func actor(c *gin.Context) model.Actor {
	value, _ := c.Get("actor")
	user, _ := value.(model.Actor)

	return user
}

// decode 解码 JSON 请求并拒绝无效或多余的数据。
func decode(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.UseNumber()
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		Failure(c, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: "无效的 JSON 请求"})
		return false
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		Failure(
			c,
			&apperror.Error{
				Status: http.StatusUnprocessableEntity,
				Detail: "请求只能包含一个 JSON 值",
			},
		)
		return false
	}

	trimmed := bytes.TrimSpace(raw)
	kind := reflect.ValueOf(target).Elem().Kind()
	if len(trimmed) == 0 || (kind == reflect.Struct && trimmed[0] != '{') || (kind == reflect.Slice && trimmed[0] != '[') {
		Failure(c, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: "请求 JSON 类型无效"})
		return false
	}

	binder := json.NewDecoder(bytes.NewReader(raw))
	binder.UseNumber()
	if err := binder.Decode(target); err != nil {
		Failure(c, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: "无效的 JSON 请求"})
		return false
	}

	return true
}

// bindInput 绑定请求参数并处理解析失败的情况。
func bindInput[T any](c *gin.Context) (T, bool) {
	var value T
	if !decode(c, &value) {
		return value, false
	}

	return value, true
}

// upload 读取上传文件及其元信息。
func upload(c *gin.Context, limit int64) (dto.Upload, func(), error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	if err := c.Request.ParseMultipartForm(limit); err != nil {
		return dto.Upload{}, func() {}, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: "文件过大或格式无效"}
	}

	cleanup := func() {
		_ = c.Request.MultipartForm.RemoveAll()
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		cleanup()
		return dto.Upload{}, func() {}, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: "需要 file 字段"}
	}

	return dto.Upload{
			Reader:      file,
			Name:        header.Filename,
			Size:        header.Size,
			ContentType: header.Header.Get("Content-Type"),
		}, func() {
			file.Close()
			cleanup()
		}, nil
}

// emit 向客户端发送一条服务端推送事件。
func emit(c *gin.Context) func(model.Record) {
	return func(event model.Record) {
		if c.Request.Context().Err() != nil {
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		raw, _ := json.Marshal(event)
		_, _ = c.Writer.Write(append(append([]byte("data: "), raw...), []byte("\n\n")...))
		c.Writer.Flush()
	}
}

// FinishPage 统一集合响应和分页字段，保持详情与树形响应不变。
func FinishPage(c *gin.Context, status int, value any, alreadyPaged bool) {
	pagination, err := dto.ParsePagination(c.Request.URL.Query())
	if err != nil {
		Failure(c, &apperror.Error{Status: http.StatusUnprocessableEntity, Detail: err.Error()})
		return
	}

	rows := []model.Record{}
	total := 0
	var category model.Record

	switch v := value.(type) {
	case []model.Record:
		rows = v
		total = len(v)
	case model.Record:
		total = int(model.Number(v, "total"))
		category, _ = v["category"].(model.Record)

		for _, key := range []string{
			"items",
			"users",
			"categories",
			"contents",
			"types",
			"templates",
			"roles",
			"questions",
			"practices",
			"assignments",
			"repositories",
			"instances",
			"tasks",
			"submissions",
			"pods",
			"nodes",
			"deployments",
			"pvs",
			"pvcs",
		} {
			if list, ok := v[key].([]model.Record); ok {
				rows = list
				break
			}
		}
	}

	if !alreadyPaged {
		total = len(rows)
		start := pagination.Offset
		if start > len(rows) {
			start = len(rows)
		}

		end := start + pagination.Limit
		if end > len(rows) {
			end = len(rows)
		}

		rows = rows[start:end]
	}

	if rows == nil {
		rows = []model.Record{}
	}

	Success(
		c,
		status,
		dto.Page[model.Record]{
			Items:    rows,
			Total:    total,
			Offset:   pagination.Offset,
			Limit:    pagination.Limit,
			Category: category,
		},
	)
}
