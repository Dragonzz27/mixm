package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/model"
)

type KTClient struct {
	Config config.Config
	HTTP   *HTTPClient
}

func (k *KTClient) request(ctx context.Context, method, path string, payload any) (model.Record, int, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(k.Config.KTURL, "/")+path, body)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := k.HTTP.Do(req, k.Config.KTTimeout)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var data model.Record
	err = json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&data)
	return data, resp.StatusCode, err
}
func (k *KTClient) Health(ctx context.Context) (model.Record, int, error) {
	return k.request(ctx, http.MethodGet, "/health", nil)
}
func (k *KTClient) Predict(ctx context.Context, payload model.Record) (model.Record, error) {
	data, status, err := k.request(ctx, http.MethodPost, "/predict", payload)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fail(http.StatusBadGateway, "知识追踪服务请求失败")
	}
	return data, nil
}
