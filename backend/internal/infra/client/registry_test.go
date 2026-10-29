package client

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"mixm/backend/internal/config"
)

type registryTransport func(*http.Request) (*http.Response, error)

// RoundTrip 为目录测试提供无需网络监听的 Registry 响应。
func (f registryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// TestRegistryCatalog 验证目录分页、认证、命名空间仓库以及错误响应不会产生部分结果。
func TestRegistryCatalog(t *testing.T) {
	for _, test := range []struct {
		name, body, link string
		status           int
		want             []string
		wantError        bool
	}{
		{
			name: "relative pagination",
			body: `{"repositories":["team/lab","mixm"]}`,
			link: `</v2/_catalog?n=2&last=team%2Flab>; rel="next"`,
			want: []string{"mixm", "team/lab", "z-image"},
		},
		{
			name: "absolute pagination",
			body: `{"repositories":["team/lab","mixm"]}`,
			link: `<http://registry.invalid/v2/_catalog?n=2&last=team%2Flab>; rel="next"`,
			want: []string{"mixm", "team/lab", "z-image"},
		},
		{
			name: "query pagination",
			body: `{"repositories":["team/lab","mixm"]}`,
			link: `<?n=2&last=team%2Flab>; rel="next"`,
			want: []string{"mixm", "team/lab", "z-image"},
		},
		{
			name: "empty",
			body: `{"repositories":[]}`,
			want: []string{},
		},
		{
			name:      "unauthorized",
			body:      `{}`,
			status:    401,
			wantError: true,
		},
		{
			name:      "malformed response",
			body:      `{`,
			wantError: true,
		},
		{
			name:      "invalid name",
			body:      `{"repositories":["../private"]}`,
			wantError: true,
		},
		{
			name:      "external link",
			body:      `{"repositories":["mixm"]}`,
			link:      `<http://other.invalid/v2/_catalog?n=2>; rel="next"`,
			wantError: true,
		},
		{
			name:      "repeated link",
			body:      `{"repositories":["mixm"]}`,
			link:      `</v2/_catalog?n=100>; rel="next"`,
			wantError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			clients := New(config.Config{
				RegistryURL:      "http://registry.invalid",
				RegistryUser:     "catalog-user",
				RegistryPassword: "test-only",
			})
			clients.HTTP.Transport = registryTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				user, password, ok := r.BasicAuth()
				if !ok || user != "catalog-user" || password != "test-only" || r.URL.Host != "registry.invalid" {
					t.Fatal("catalog authentication sent to an unexpected destination")
				}
				if calls > 2 {
					t.Fatal("catalog pagination did not terminate")
				}
				body, link := test.body, test.link
				if calls == 2 {
					if r.URL.Query().Get("last") != "team/lab" {
						t.Fatal("pagination lost its cursor")
					}
					body, link = `{"repositories":["team/lab","z-image"]}`, ""
				}
				status := test.status
				if status == 0 {
					status = http.StatusOK
				}
				return &http.Response{
					StatusCode: status,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Link": []string{link}},
				}, nil
			})
			got, err := clients.Registry.Catalog(context.Background())
			if test.wantError {
				if err == nil || got != nil {
					t.Fatal("invalid catalog must fail without partial results")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("unexpected catalog: %v, %v", got, err)
			}
		})
	}
}

// TestRegistryTags 验证标签分页沿用同一仓库路径，空标签与查询失败明确区分。
func TestRegistryTags(t *testing.T) {
	for _, test := range []struct {
		name, body string
		paged      bool
		failNext   bool
		want       []string
		wantError  bool
	}{
		{
			name:  "all pages",
			body:  `{"name":"team/lab","tags":["latest","v0.0.1"]}`,
			paged: true,
			want:  []string{"latest", "v0.0.1", "v0.0.2"},
		},
		{
			name: "no tags",
			body: `{"name":"team/lab","tags":null}`,
			want: []string{},
		},
		{
			name:      "invalid response",
			body:      `{"repositories":["team/lab"]}`,
			wantError: true,
		},
		{
			name:      "later page unavailable",
			body:      `{"tags":["latest"]}`,
			paged:     true,
			failNext:  true,
			wantError: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			clients := New(config.Config{RegistryURL: "http://registry.invalid"})
			calls := 0
			clients.HTTP.Transport = registryTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Path != "/v2/team/lab/tags/list" || calls > 2 {
					t.Fatal("tag pagination requested an unexpected endpoint")
				}

				body, status := test.body, http.StatusOK
				header := http.Header{}
				if calls == 1 && test.paged {
					header.Set("Link", `</v2/team/lab/tags/list?n=2&last=v0.0.1>; rel="next"`)
				}
				if calls == 2 {
					body = `{"name":"team/lab","tags":["v0.0.1","v0.0.2"]}`
					if test.failNext {
						status = http.StatusServiceUnavailable
					}
				}

				return &http.Response{
					StatusCode: status,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     header,
				}, nil
			})
			got, err := clients.Registry.Tags(context.Background(), "team/lab")
			if test.wantError {
				if err == nil || got != nil {
					t.Fatal("invalid tag query must not appear as an empty or partial tag list")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, test.want) {
				t.Fatalf("unexpected tags: %v, %v", got, err)
			}
		})
	}
}
