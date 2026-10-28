package client

import (
	"context"
	"net/http"
	"strings"
	"time"

	"mixm/backend/internal/config"
)

type RegistryClient struct {
	Config config.Config
	HTTP   *HTTPClient
}

func (s *RegistryClient) Request(ctx context.Context, method, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(s.Config.RegistryURL, "/")+"/v2/"+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.Config.RegistryUser, s.Config.RegistryPassword)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
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
