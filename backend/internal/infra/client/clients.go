package client

import (
	"net/http"
	"time"

	"mixm-edu/backend/internal/apperror"
	"mixm-edu/backend/internal/config"
)

// Clients is constructed once and injects independently configured external adapters.
type Clients struct {
	HTTP     *HTTPClient
	LLM      *LLMClient
	Registry *RegistryClient
	Objects  *ObjectStore
	KT       *KTClient
}
type HTTPClient struct{ Transport http.RoundTripper }

func (h *HTTPClient) Do(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{Transport: h.Transport, Timeout: timeout}).Do(req)
}
func New(c config.Config) *Clients {
	transport := &HTTPClient{}
	return &Clients{HTTP: transport, LLM: &LLMClient{Config: c, HTTP: transport}, Registry: &RegistryClient{Config: c, HTTP: transport}, Objects: &ObjectStore{Config: c}, KT: &KTClient{Config: c, HTTP: transport}}
}
func fail(status int, message string) error { return &apperror.Error{Status: status, Detail: message} }
