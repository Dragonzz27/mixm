package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"mixm/backend/internal/config"
)

type RegistryClient struct {
	Config config.Config
	HTTP   *HTTPClient
}

var repositoryName = regexp.MustCompile(`^[a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*)*$`)

// Catalog 按 Registry 的分页链接读取完整仓库目录，查询失败时不返回部分结果。
func (s *RegistryClient) Catalog(ctx context.Context) ([]string, error) {
	names, err := s.listStrings(ctx, "_catalog", "repositories")
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if len(name) > 255 || !repositoryName.MatchString(name) {
			return nil, fail(http.StatusBadGateway, "Registry 返回了无效的仓库名称")
		}
	}

	return names, nil
}

// Tags 读取仓库的全部标签，包含 Registry 分页返回的后续标签。
func (s *RegistryClient) Tags(ctx context.Context, name string) ([]string, error) {
	if len(name) > 255 || !repositoryName.MatchString(name) {
		return nil, fail(http.StatusUnprocessableEntity, "镜像仓库名称无效")
	}

	return s.listStrings(ctx, name+"/tags/list", "tags")
}

// listStrings 读取 Registry 字符串集合，按 Link 分页并拒绝不完整或异常的响应。
func (s *RegistryClient) listStrings(ctx context.Context, endpoint, field string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	base, err := url.Parse(strings.TrimSuffix(s.Config.RegistryURL, "/") + "/v2/" + endpoint)
	if err != nil {
		return nil, err
	}

	path := endpoint + "?n=100"
	visited := map[string]bool{}
	names := map[string]bool{}

	for path != "" {
		if visited[path] {
			return nil, fail(http.StatusBadGateway, "Registry 分页链接重复")
		}

		visited[path] = true
		resp, err := s.Request(ctx, http.MethodGet, path)
		if err != nil {
			return nil, err
		}

		var data map[string]json.RawMessage
		err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&data)
		resp.Body.Close()
		if err != nil || data[field] == nil {
			return nil, fail(http.StatusBadGateway, "Registry 列表响应无效")
		}

		var values []string
		if err := json.Unmarshal(data[field], &values); err != nil {
			return nil, fail(http.StatusBadGateway, "Registry 列表响应无效")
		}

		for _, value := range values {
			if value == "" {
				return nil, fail(http.StatusBadGateway, "Registry 列表中存在空值")
			}
			names[value] = true
		}

		// 分页链接只允许访问当前 Registry 的同一端点，避免向其他地址发送认证信息。
		current := *base
		current.RawQuery = strings.TrimPrefix(path, endpoint+"?")
		path = ""
		for _, header := range resp.Header.Values("Link") {
			for _, link := range strings.Split(header, ",") {
				parts := strings.Split(link, ";")
				isNext := false
				for _, parameter := range parts[1:] {
					if strings.TrimSpace(parameter) == `rel="next"` || strings.TrimSpace(parameter) == "rel=next" {
						isNext = true
					}
				}

				if !isNext {
					continue
				}

				target, err := url.Parse(strings.Trim(strings.TrimSpace(parts[0]), "<>"))
				if err != nil {
					return nil, fail(http.StatusBadGateway, "Registry 分页链接无效")
				}

				next := current.ResolveReference(target)
				if next.Scheme != base.Scheme || next.Host != base.Host || next.Path != base.Path ||
					next.User != nil || next.Fragment != "" {
					return nil, fail(http.StatusBadGateway, "Registry 分页链接无效")
				}

				path = endpoint + "?" + next.RawQuery
			}
		}
	}

	out := make([]string, 0, len(names))

	for name := range names {
		out = append(out, name)
	}

	sort.Strings(out)

	return out, nil
}

// Request 向镜像仓库发送带认证的 API 请求。
func (s *RegistryClient) Request(ctx context.Context, method, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(s.Config.RegistryURL, "/")+"/v2/"+path, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.Config.RegistryUser, s.Config.RegistryPassword)
	req.Header.Set(
		"Accept",
		"application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json",
	)
	resp, err := s.HTTP.Do(req, 30*time.Second)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		if resp.StatusCode == 404 {
			return nil, fail(http.StatusNotFound, "镜像或标签不存在")
		}

		return nil, fail(http.StatusBadGateway, "镜像仓库请求失败")
	}

	return resp, nil
}
