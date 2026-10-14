package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"mixm-edu/backend/internal/store"
)

func (s *Server) objectClient() (*minio.Client, error) {
	u, e := url.Parse(s.Config.RustEndpoint)
	if e != nil {
		return nil, e
	}
	return minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(s.Config.RustKey, s.Config.RustSecret, ""), Secure: u.Scheme == "https"})
}
func (s *Server) fileRoutes() {
	s.route("POST /api/assignments/upload", func(w http.ResponseWriter, r *http.Request) (any, error) {
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
		if e := r.ParseMultipartForm(50 << 20); e != nil {
			return nil, fail(422, "文件过大或格式无效")
		}
		defer r.MultipartForm.RemoveAll()
		f, h, e := r.FormFile("file")
		if e != nil {
			return nil, fail(422, "需要 file 字段")
		}
		defer f.Close()
		ext := strings.ToLower(filepath.Ext(h.Filename))
		if !map[string]bool{".doc": true, ".docx": true, ".pdf": true, ".ppt": true, ".pptx": true, ".txt": true}[ext] {
			return nil, fail(422, "文件类型不支持")
		}
		if e = os.MkdirAll(s.Config.UploadDir, 0755); e != nil {
			return nil, e
		}
		name := uuid.NewString() + ext
		p := filepath.Join(s.Config.UploadDir, name)
		out, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return nil, e
		}
		size, e := io.Copy(out, f)
		ce := out.Close()
		if e != nil {
			os.Remove(p)
			return nil, e
		}
		if ce != nil {
			return nil, ce
		}
		return store.Row{"file_path": p, "file_name": h.Filename, "file_size": size, "file_url": "/api/assignments/file/" + name}, nil
	})
	s.route("GET /api/assignments/file/{filename}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		name := r.PathValue("filename")
		if name != filepath.Base(name) || name == "." || name == ".." {
			return nil, fail(404, "文件不存在")
		}
		p := filepath.Join(s.Config.UploadDir, name)
		if _, e := os.Stat(p); e != nil {
			return nil, fail(404, "文件不存在")
		}
		http.ServeFile(w, r, p)
		return nil, nil
	}, true)
	s.route("POST /api/contents/upload", func(w http.ResponseWriter, r *http.Request) (any, error) {
		if e := s.permission(r, "Contents"); e != nil {
			return nil, e
		}
		r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
		if e := r.ParseMultipartForm(100 << 20); e != nil {
			return nil, fail(422, "文件过大或格式无效")
		}
		defer r.MultipartForm.RemoveAll()
		f, h, e := r.FormFile("file")
		if e != nil {
			return nil, fail(422, "需要 file 字段")
		}
		defer f.Close()
		c, e := s.objectClient()
		if e != nil {
			return nil, e
		}
		exists, e := c.BucketExists(r.Context(), s.Config.RustBucket)
		if e != nil {
			return nil, e
		}
		if !exists {
			if e = c.MakeBucket(r.Context(), s.Config.RustBucket, minio.MakeBucketOptions{}); e != nil {
				return nil, e
			}
		}
		key := uuid.NewString() + filepath.Ext(filepath.Base(h.Filename))
		_, e = c.PutObject(r.Context(), s.Config.RustBucket, key, f, h.Size, minio.PutObjectOptions{ContentType: h.Header.Get("Content-Type")})
		if e != nil {
			return nil, e
		}
		return store.Row{"key": key, "file_name": h.Filename, "filename": h.Filename, "file_size": h.Size, "url": "/api/contents/file/" + key}, nil
	})
	s.route("GET /api/contents/file/{key...}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		c, e := s.objectClient()
		if e != nil {
			return nil, e
		}
		obj, e := c.GetObject(r.Context(), s.Config.RustBucket, r.PathValue("key"), minio.GetObjectOptions{})
		if e != nil {
			return nil, e
		}
		defer obj.Close()
		info, e := obj.Stat()
		if e != nil {
			return nil, fail(404, "文件不存在")
		}
		w.Header().Set("Content-Type", info.ContentType)
		w.Header().Set("Content-Length", fmt.Sprint(info.Size))
		_, e = io.Copy(w, obj)
		return nil, e
	}, true)
}
