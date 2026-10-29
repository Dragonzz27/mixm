package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
	"mixm/backend/internal/testutil"
)

// TestImageRegistrySync 验证发现、权限、重复同步、元数据保留以及 Registry 故障时的列表行为。
func TestImageRegistrySync(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user := func(name string, superuser bool) model.Record {
		t.Helper()
		v, err := repository.Write(ctx, db, "users", "", model.Record{
			"username":        name,
			"hashed_password": "unused",
			"is_superuser":    superuser,
		})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	admin := user("sync-admin", true)
	owner := user("existing-owner", false)
	member := user("ordinary-member", false)
	existing, err := repository.Write(ctx, db, "image_repositories", "", model.Record{
		"name":      "private-lab",
		"user_id":   owner["id"],
		"is_public": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	stale, err := repository.Write(ctx, db, "image_repositories", "", model.Record{
		"name":      "not-in-catalog",
		"user_id":   owner["id"],
		"is_public": false,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Registry 请求只访问测试服务器，不修改真实镜像或业务数据库。
	var mode, requests atomic.Int32
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodGet {
			t.Error("sync must not modify the Registry")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if mode.Load() == 1 || (mode.Load() == 2 && r.URL.Query().Has("last")) {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/v2/_catalog" {
			fmt.Fprint(w, `{"tags":["v0.0.1","latest"]}`)
			return
		}
		if mode.Load() == 3 {
			fmt.Fprint(w, `{"repositories":[]}`)
		} else if r.URL.Query().Has("last") {
			fmt.Fprint(w, `{"repositories":["team/gpu-lab","mixm-codeserver"]}`)
		} else {
			w.Header().Set("Link", `</v2/_catalog?n=2&last=private-lab>; rel="next"`)
			if mode.Load() == 2 {
				fmt.Fprint(w, `{"repositories":["partial-only"]}`)
			} else {
				fmt.Fprint(w, `{"repositories":["private-lab","mixm-codeserver"]}`)
			}
		}
	}))
	defer registry.Close()
	settings := config.Config{
		Secret:      "sync-test-secret",
		RegistryURL: registry.URL,
	}
	router := New(settings, db, nil, nil)
	token := func(user model.Record) string {
		t.Helper()
		value, err := auth.Sign(settings.Secret, model.String(user, "username"), "access", time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	call := func(method, path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	data := func(response *httptest.ResponseRecorder, status int) model.Record {
		t.Helper()
		if response.Code != status {
			t.Fatalf("request returned %d, want %d: %s", response.Code, status, response.Body.String())
		}
		var envelope model.Record
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return model.Object(envelope["data"])
	}
	adminToken := token(admin)
	data(call(http.MethodPost, "/api/images/sync", token(member)), http.StatusForbidden)
	if requests.Load() != 0 {
		t.Fatal("ordinary users must not read the Registry catalog")
	}

	first := data(call(http.MethodPost, "/api/images/sync", adminToken), http.StatusOK)
	if model.Number(first, "total") != 3 || model.Number(first, "created") != 2 {
		t.Fatalf("unexpected initial sync: %v", first)
	}
	second := data(call(http.MethodPost, "/api/images/sync", adminToken), http.StatusOK)
	if model.Number(second, "created") != 0 {
		t.Fatal("repeated sync created duplicate records")
	}
	for _, before := range []model.Record{existing, stale} {
		after, err := repository.Get(ctx, db, "image_repositories", model.String(before, "id"))
		if err != nil || model.Bool(after, "is_public") || after["user_id"] != before["user_id"] || after["updated_at"] != nil {
			t.Fatal("sync overwrote or removed existing repository metadata")
		}
	}
	page := data(call(http.MethodGet, "/api/images", token(member)), http.StatusOK)
	if model.Number(page, "total") != 2 {
		t.Fatal("new repositories must be public without exposing existing private repositories")
	}
	for _, item := range model.Objects(page["items"]) {
		if item["user_id"] != admin["id"] || model.Number(item, "tags_count") != 2 {
			t.Fatal("discovered repository ownership or live tags incorrect")
		}
	}

	// 同时打开多个管理员页面也不应产生重复记录。
	var workers sync.WaitGroup
	for range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			response := call(http.MethodPost, "/api/images/sync", adminToken)
			if response.Code != http.StatusOK {
				t.Errorf("concurrent sync failed: %d", response.Code)
			}
		}()
	}
	workers.Wait()

	mode.Store(2)
	data(call(http.MethodPost, "/api/images/sync", adminToken), http.StatusBadGateway)
	rows, err := repository.Query(ctx, db, "SELECT id FROM image_repositories WHERE name=$1", "partial-only")
	if err != nil || len(rows) != 0 {
		t.Fatal("failed pagination wrote a partial catalog")
	}
	mode.Store(1)
	data(call(http.MethodPost, "/api/images/sync", adminToken), http.StatusBadGateway)
	page = data(call(http.MethodGet, "/api/images", adminToken), http.StatusOK)
	if model.Number(page, "total") != 4 {
		t.Fatal("registered repositories disappeared during Registry outage")
	}
	for _, item := range model.Objects(page["items"]) {
		if item["tags_count"] != nil {
			t.Fatal("unavailable tag counts must remain unknown")
		}
	}
	mode.Store(3)
	empty := data(call(http.MethodPost, "/api/images/sync", adminToken), http.StatusOK)
	if model.Number(empty, "total") != 0 || model.Number(empty, "created") != 0 {
		t.Fatal("empty Registry catalog must sync successfully")
	}
}

// TestImageTagsRoundTrip 验证标签在仓库详情、统计与分页接口中一致，查询错误不会被伪装为空集合。
func TestImageTagsRoundTrip(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	user, err := repository.Write(ctx, db, "users", "", model.Record{
		"username":        "tag-owner",
		"hashed_password": "unused",
	})
	if err != nil {
		t.Fatal(err)
	}
	image, err := repository.Write(ctx, db, "image_repositories", "", model.Record{
		"name":      "team/mixm",
		"user_id":   user["id"],
		"is_public": true,
	})
	if err != nil {
		t.Fatal(err)
	}

	var unavailable atomic.Bool
	registry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v2/team/mixm/tags/list" {
			if unavailable.Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			if r.URL.Query().Has("last") {
				fmt.Fprint(w, `{"tags":["v0.0.1"]}`)
			} else {
				w.Header().Set("Link", `</v2/team/mixm/tags/list?n=1&last=latest>; rel="next"`)
				fmt.Fprint(w, `{"tags":["latest"]}`)
			}
			return
		}
		fmt.Fprint(w, `{"config":{"size":10},"layers":[{"size":20}]}`)
	}))
	defer registry.Close()
	settings := config.Config{
		Secret:         "tag-test-secret",
		RegistryURL:    registry.URL,
		RegistryPrefix: "registry.test:55000",
	}
	router := New(settings, db, nil, nil)
	token, err := auth.Sign(settings.Secret, model.String(user, "username"), "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	call := func(path string, status int) model.Record {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != status {
			t.Fatalf("%s returned %d, want %d", path, response.Code, status)
		}
		var envelope model.Record
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return model.Object(envelope["data"])
	}
	path := "/api/images/" + model.String(image, "id")
	detail := call(path, http.StatusOK)
	tags := model.Objects(detail["tags"])
	if len(tags) != 2 || model.Number(detail, "tags_count") != 2 {
		t.Fatal("repository detail lost paginated tags")
	}
	for _, tag := range tags {
		if model.Number(tag, "size") != 30 || model.String(tag, "full_name") != "registry.test:55000/team/mixm:"+model.String(tag, "name") {
			t.Fatal("tag metadata or image reference changed")
		}
	}
	page := call(path+"/tags?offset=1&limit=1", http.StatusOK)
	if model.Number(page, "total") != 2 || len(model.Objects(page["items"])) != 1 ||
		model.String(model.Objects(page["items"])[0], "name") != "v0.0.1" {
		t.Fatal("tag API pagination does not match repository detail")
	}
	list := call("/api/images", http.StatusOK)
	if model.Number(model.Objects(list["items"])[0], "tags_count") != 2 {
		t.Fatal("repository count omitted later tag pages")
	}

	// 回归原先详情接口覆盖标签查询错误、返回成功空列表的问题。
	unavailable.Store(true)
	call(path, http.StatusBadGateway)
	call(path+"/tags", http.StatusBadGateway)
}
