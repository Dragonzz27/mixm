package client

import (
	"net/http"
	"time"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/config"
)

// Clients 集中创建外部适配器，并共享可替换的 HTTP 传输。
type Clients struct {
	HTTP     *HTTPClient
	LLM      *LLMClient
	Registry *RegistryClient
	Objects  *ObjectStore
}

type HTTPClient struct{ Transport http.RoundTripper }

// Do 使用指定超时时间发送 HTTP 请求。
func (h *HTTPClient) Do(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{Transport: h.Transport, Timeout: timeout}).Do(req)
}

// New 创建共享 HTTP 传输及独立的外部服务适配器。
func New(c config.Config) *Clients {
	transport := &HTTPClient{}
	return &Clients{
		HTTP:     transport,
		LLM:      &LLMClient{Config: c, HTTP: transport},
		Registry: &RegistryClient{Config: c, HTTP: transport},
		Objects:  &ObjectStore{Config: c},
	}
}

// fail 构造携带 HTTP 状态和公开消息的业务错误。
func fail(status int, message string) error {
	return &apperror.Error{Status: status, Detail: message}
}
