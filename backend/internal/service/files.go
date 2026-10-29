package service

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
)

type Files struct{ *Core }

// AssignmentUpload 保存作业附件并返回文件信息。
func (s *Files) AssignmentUpload(upload dto.Upload) (any, error) {
	f := upload.Reader
	var err error

	ext := strings.ToLower(filepath.Ext(upload.Name))
	if !map[string]bool{
		".doc":  true,
		".docx": true,
		".pdf":  true,
		".ppt":  true,
		".pptx": true,
		".txt":  true,
	}[ext] {
		return nil, fail(http.StatusUnprocessableEntity, "文件类型不支持")
	}

	if err = os.MkdirAll(s.Config.UploadDir, 0755); err != nil {
		return nil, err
	}

	name := uuid.NewString() + ext
	p := filepath.Join(s.Config.UploadDir, name)
	out, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	size, err := io.Copy(out, f)
	ce := out.Close()
	if err != nil {
		os.Remove(p)
		return nil, err
	}

	if ce != nil {
		return nil, ce
	}

	return model.Record{
		"file_path": p,
		"file_name": upload.Name,
		"file_size": size,
		"file_url":  "/api/assignments/files/" + name,
	}, nil
}

// AssignmentDownload 校验附件路径并返回作业文件。
func (s *Files) AssignmentDownload(filename string) (any, error) {
	name := filename
	if name != filepath.Base(name) || name == "." || name == ".." {
		return nil, fail(http.StatusNotFound, "文件不存在")
	}

	p := filepath.Join(s.Config.UploadDir, name)
	if _, err := os.Stat(p); err != nil {
		return nil, fail(http.StatusNotFound, "文件不存在")
	}

	return dto.File{Path: p}, nil
}

// ContentUpload 上传教学内容附件到对象存储。
func (s *Files) ContentUpload(ctx context.Context, actor model.Actor, upload dto.Upload) (any, error) {
	if err := s.permission(ctx, actor, "Contents"); err != nil {
		return nil, err
	}

	f := upload.Reader
	var err error

	c, err := s.Clients.Objects.Client()
	if err != nil {
		return nil, err
	}

	exists, err := c.BucketExists(ctx, s.Config.RustBucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err = c.MakeBucket(ctx, s.Config.RustBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	key := uuid.NewString() + filepath.Ext(filepath.Base(upload.Name))
	_, err = c.PutObject(ctx, s.Config.RustBucket, key, f, upload.Size, minio.PutObjectOptions{ContentType: upload.ContentType})
	if err != nil {
		return nil, err
	}

	return model.Record{
		"key":       key,
		"file_name": upload.Name,
		"filename":  upload.Name,
		"file_size": upload.Size,
		"url":       "/api/contents/files/" + key,
	}, nil
}

// ContentDownload 读取对象存储中的内容附件。
func (s *Files) ContentDownload(ctx context.Context, key string) (any, error) {
	c, err := s.Clients.Objects.Client()
	if err != nil {
		return nil, err
	}

	obj, err := c.GetObject(ctx, s.Config.RustBucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, fail(http.StatusNotFound, "文件不存在")
	}

	return dto.File{
		Reader:      obj,
		Size:        info.Size,
		ContentType: info.ContentType,
	}, nil
}
