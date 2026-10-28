package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/dto"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Images struct{ *Core }

func (s *Images) image(ctx context.Context, actor model.Actor, id string) (model.Record, error) {
	v, err := repository.Get(ctx, s.DB, "image_repositories", id)
	if err != nil {
		return nil, err
	}
	if !model.Bool(v, "is_public") && v["user_id"] != actor.ID {
		if err = s.permission(ctx, actor, "AdminImageRepositories"); err != nil {
			return nil, err
		}
	}
	return v, nil
}
func (s *Images) tags(ctx context.Context, v model.Record, detailed ...bool) ([]model.Record, error) {
	resp, err := s.Clients.Registry.Request(ctx, "GET", model.String(v, "name")+"/tags/list")
	if err != nil {
		var he *apperror.Error
		if errors.As(err, &he) && he.Status == 404 {
			return []model.Record{}, nil
		}
		return nil, err
	}
	defer resp.Body.Close()
	var data struct {
		Tags []string `json:"tags"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&data); err != nil {
		return nil, err
	}
	out := []model.Record{}
	for _, t := range data.Tags {
		item := model.Record{"name": t, "tag": t, "size": 0, "pushed_at": nil, "created_at": nil, "full_name": s.Config.RegistryPrefix + "/" + model.String(v, "name") + ":" + t}
		if len(detailed) == 0 || detailed[0] {
			resp, err := s.Clients.Registry.Request(ctx, "GET", model.String(v, "name")+"/manifests/"+url.PathEscape(t))
			if err == nil {
				var manifest model.Record
				err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&manifest)
				resp.Body.Close()
				if err == nil {
					size := model.Number(model.Object(manifest["config"]), "size")
					for _, layer := range model.Objects(manifest["layers"]) {
						size += model.Number(layer, "size")
					}
					item["size"] = size
				}
			}
		}
		out = append(out, item)
	}
	return out, nil
}
func (s *Images) deleteTag(ctx context.Context, v model.Record, tag string) error {
	p := model.String(v, "name") + "/manifests/" + url.PathEscape(tag)
	resp, err := s.Clients.Registry.Request(ctx, "HEAD", p)
	if err != nil {
		return err
	}
	resp.Body.Close()
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fmt.Errorf("registry did not return manifest digest")
	}
	resp, err = s.Clients.Registry.Request(ctx, "DELETE", model.String(v, "name")+"/manifests/"+url.PathEscape(digest))
	if resp != nil {
		resp.Body.Close()
	}
	return err
}
func (s *Images) List(ctx context.Context, actor model.Actor, query url.Values, mine bool) (any, error) {
	extra := "(is_public=true OR user_id=$1)"
	if mine {
		extra = "user_id=$1"
	}
	if actor.IsSuperuser && !mine {
		extra = "true OR user_id=$1"
	}
	rs, total, err := s.list(ctx, actor, query, "image_repositories", extra, []any{actor.ID})
	if err != nil {
		return nil, err
	}
	for _, v := range rs {
		u, err := repository.Get(ctx, s.DB, "users", model.String(v, "user_id"))
		if err != nil {
			return nil, err
		}
		v["user_info"] = model.Record{"id": u["id"], "username": u["username"]}
		tags, err := s.tags(ctx, v, false)
		if err != nil {
			return nil, err
		}
		v["tags_count"] = len(tags)
	}
	return model.Record{"total": total, "repositories": rs}, nil
}
func (s *Images) Create(ctx context.Context, actor model.Actor, input dto.ImageInput) (any, error) {
	v := input.Fields()
	var err error
	if err = require(v, "name"); err != nil {
		return nil, err
	}
	if strings.Contains(model.String(v, "name"), "..") || strings.ContainsAny(model.String(v, "name"), "?#") {
		return nil, fail(http.StatusUnprocessableEntity, "镜像仓库名称无效")
	}
	v["user_id"] = actor.ID
	out, err := repository.Write(ctx, s.DB, "image_repositories", "", v)
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (s *Images) Detail(ctx context.Context, actor model.Actor, id string) (any, error) {
	v, err := s.image(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	v["tags"], err = s.tags(ctx, v)
	v["tags_count"] = len(model.Objects(v["tags"]))
	u, err := repository.Get(ctx, s.DB, "users", model.String(v, "user_id"))
	if err != nil {
		return nil, err
	}
	v["user_info"] = model.Record{"id": u["id"], "username": u["username"]}
	return v, err
}
func (s *Images) Tags(ctx context.Context, actor model.Actor, id string) (any, error) {
	v, err := s.image(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	return s.tags(ctx, v)
}
func (s *Images) Update(ctx context.Context, actor model.Actor, input dto.ImageInput, id string) (any, error) {
	v, err := s.image(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if v["user_id"] != actor.ID {
		if err = s.permission(ctx, actor, "AdminImageRepositories"); err != nil {
			return nil, err
		}
	}
	data := input.Fields()
	delete(data, "user_id")
	return repository.Write(ctx, s.DB, "image_repositories", id, data)
}
func (s *Images) Delete(ctx context.Context, actor model.Actor, id string) (any, error) {
	v, err := s.image(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if v["user_id"] != actor.ID {
		if err = s.permission(ctx, actor, "AdminImageRepositories"); err != nil {
			return nil, err
		}
	}
	tags, err := s.tags(ctx, v)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		if err = s.deleteTag(ctx, v, model.String(t, "name")); err != nil {
			return nil, err
		}
	}
	return model.Record{}, repository.Delete(ctx, s.DB, "image_repositories", id)
}
func (s *Images) DeleteTag(ctx context.Context, actor model.Actor, id string, tag string) (any, error) {
	v, err := s.image(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if v["user_id"] != actor.ID {
		if err = s.permission(ctx, actor, "AdminImageRepositories"); err != nil {
			return nil, err
		}
	}
	return model.Record{}, s.deleteTag(ctx, v, tag)
}
