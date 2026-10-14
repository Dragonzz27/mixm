package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mixm-edu/backend/internal/store"
)

func (s *Server) registry(r *http.Request, method, path string) (*http.Response, error) {
	req, e := http.NewRequestWithContext(r.Context(), method, strings.TrimSuffix(s.Config.RegistryURL, "/")+"/v2/"+path, nil)
	if e != nil {
		return nil, e
	}
	req.SetBasicAuth(s.Config.RegistryUser, s.Config.RegistryPassword)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
	resp, e := s.do(req, 30*time.Second)
	if e != nil {
		return nil, e
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		if resp.StatusCode == 404 {
			return nil, fail(404, "镜像或标签不存在")
		}
		return nil, fail(502, "镜像仓库请求失败")
	}
	return resp, nil
}
func (s *Server) image(r *http.Request) (store.Row, error) {
	v, e := store.Get(r.Context(), s.DB, "image_repositories", r.PathValue("id"))
	if e != nil {
		return nil, e
	}
	if !store.Bool(v, "is_public") && v["user_id"] != user(r)["id"] {
		if e = s.permission(r, "AdminImageRepositories"); e != nil {
			return nil, e
		}
	}
	return v, nil
}
func (s *Server) tags(r *http.Request, v store.Row, detailed ...bool) ([]store.Row, error) {
	resp, e := s.registry(r, "GET", store.String(v, "name")+"/tags/list")
	if e != nil {
		var he *Error
		if errors.As(e, &he) && he.Status == 404 {
			return []store.Row{}, nil
		}
		return nil, e
	}
	defer resp.Body.Close()
	var data struct {
		Tags []string `json:"tags"`
	}
	if e = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&data); e != nil {
		return nil, e
	}
	out := []store.Row{}
	for _, t := range data.Tags {
		item := store.Row{"name": t, "tag": t, "size": 0, "pushed_at": nil, "created_at": nil, "full_name": s.Config.RegistryPrefix + "/" + store.String(v, "name") + ":" + t}
		if len(detailed) == 0 || detailed[0] {
			resp, err := s.registry(r, "GET", store.String(v, "name")+"/manifests/"+url.PathEscape(t))
			if err == nil {
				var manifest store.Row
				err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&manifest)
				resp.Body.Close()
				if err == nil {
					size := store.Number(store.Object(manifest["config"]), "size")
					for _, layer := range store.Objects(manifest["layers"]) {
						size += store.Number(layer, "size")
					}
					item["size"] = size
				}
			}
		}
		out = append(out, item)
	}
	return out, nil
}
func (s *Server) imageRoutes() {
	for _, p := range []string{"/api/images", "/api/images/me"} {
		s.route("GET "+p, func(w http.ResponseWriter, r *http.Request) (any, error) {
			extra := "(is_public=true OR user_id=$1)"
			if r.URL.Path == "/api/images/me" {
				extra = "user_id=$1"
			}
			if store.Bool(user(r), "is_superuser") && r.URL.Path == "/api/images" {
				extra = "true OR user_id=$1"
			}
			rs, total, e := s.list(r, "image_repositories", extra, []any{user(r)["id"]})
			if e != nil {
				return nil, e
			}
			for _, v := range rs {
				u, e := store.Get(r.Context(), s.DB, "users", store.String(v, "user_id"))
				if e != nil {
					return nil, e
				}
				v["user_info"] = store.Row{"id": u["id"], "username": u["username"]}
				tags, e := s.tags(r, v, false)
				if e != nil {
					return nil, e
				}
				v["tags_count"] = len(tags)
			}
			return store.Row{"total": total, "repositories": rs}, nil
		})
	}
	s.route("POST /api/images", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := body(r)
		if e != nil {
			return nil, e
		}
		if e = require(v, "name"); e != nil {
			return nil, e
		}
		if strings.Contains(store.String(v, "name"), "..") || strings.ContainsAny(store.String(v, "name"), "?#") {
			return nil, fail(422, "镜像仓库名称无效")
		}
		v["user_id"] = user(r)["id"]
		out, e := store.Write(r.Context(), s.DB, "image_repositories", "", v)
		if e != nil {
			return nil, e
		}
		respond(w, 201, out)
		return nil, nil
	})
	s.route("GET /api/images/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.image(r)
		if e != nil {
			return nil, e
		}
		v["tags"], e = s.tags(r, v)
		v["tags_count"] = len(store.Objects(v["tags"]))
		u, err := store.Get(r.Context(), s.DB, "users", store.String(v, "user_id"))
		if err != nil {
			return nil, err
		}
		v["user_info"] = store.Row{"id": u["id"], "username": u["username"]}
		return v, e
	})
	s.route("GET /api/images/{id}/tags", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.image(r)
		if e != nil {
			return nil, e
		}
		return s.tags(r, v)
	})
	s.route("PUT /api/images/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.image(r)
		if e != nil {
			return nil, e
		}
		if v["user_id"] != user(r)["id"] {
			if e = s.permission(r, "AdminImageRepositories"); e != nil {
				return nil, e
			}
		}
		data, e := body(r)
		if e != nil {
			return nil, e
		}
		delete(data, "user_id")
		return store.Write(r.Context(), s.DB, "image_repositories", r.PathValue("id"), data)
	})
	s.route("DELETE /api/images/{id}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.image(r)
		if e != nil {
			return nil, e
		}
		if v["user_id"] != user(r)["id"] {
			if e = s.permission(r, "AdminImageRepositories"); e != nil {
				return nil, e
			}
		}
		tags, e := s.tags(r, v)
		if e != nil {
			return nil, e
		}
		for _, t := range tags {
			if e = s.deleteTag(r, v, store.String(t, "name")); e != nil {
				return nil, e
			}
		}
		return store.Row{}, store.Delete(r.Context(), s.DB, "image_repositories", r.PathValue("id"))
	})
	s.route("DELETE /api/images/{id}/tags/{tag}", func(w http.ResponseWriter, r *http.Request) (any, error) {
		v, e := s.image(r)
		if e != nil {
			return nil, e
		}
		if v["user_id"] != user(r)["id"] {
			if e = s.permission(r, "AdminImageRepositories"); e != nil {
				return nil, e
			}
		}
		return store.Row{}, s.deleteTag(r, v, r.PathValue("tag"))
	})
}
func (s *Server) deleteTag(r *http.Request, v store.Row, tag string) error {
	p := store.String(v, "name") + "/manifests/" + url.PathEscape(tag)
	resp, e := s.registry(r, "HEAD", p)
	if e != nil {
		return e
	}
	resp.Body.Close()
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fmt.Errorf("registry did not return manifest digest")
	}
	resp, e = s.registry(r, "DELETE", store.String(v, "name")+"/manifests/"+url.PathEscape(digest))
	if resp != nil {
		resp.Body.Close()
	}
	return e
}
