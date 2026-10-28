package testutil

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"mixm/backend/internal/repository"
)

func Database(t *testing.T) *repository.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	root, err := repository.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	name := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = root.Exec(ctx, "CREATE SCHEMA "+name); err != nil {
		root.Close()
		t.Fatal(err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	params := parsed.Query()
	params.Set("search_path", name)
	parsed.RawQuery = params.Encode()
	db, err := repository.Open(ctx, parsed.String())
	if err != nil {
		root.Exec(ctx, "DROP SCHEMA "+name+" CASCADE")
		root.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := root.Exec(cleanup, "DROP SCHEMA "+name+" CASCADE"); err != nil {
			t.Error(err)
		}
		root.Close()
	})
	if err = repository.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db
}
