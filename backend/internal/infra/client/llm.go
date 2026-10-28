package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
)

type LLMClient struct {
	Config config.Config
	HTTP   *HTTPClient
}

func (s *LLMClient) Request(ctx context.Context, prompt string, stream bool) (*http.Response, error) {
	if s.Config.LLMKey == "" {
		return nil, fail(http.StatusServiceUnavailable, "LLM_API_KEY 未配置")
	}
	payload := model.Record{"model": s.Config.LLMModel, "messages": []model.Record{{"role": "user", "content": prompt}}, "response_format": model.Record{"type": "json_object"}, "stream": stream}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimSuffix(s.Config.LLMURL, "/")+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.Config.LLMKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.HTTP.Do(req, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fail(http.StatusBadGateway, "LLM 请求失败")
	}
	return resp, nil
}
