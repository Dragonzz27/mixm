package testutil

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"mixm-edu/backend/internal/store"
)

func Database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx := context.Background()
	root, e := store.Open(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	name := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, e = root.Exec(ctx, "CREATE SCHEMA "+name); e != nil {
		root.Close()
		t.Fatal(e)
	}
	cfg, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = name
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, e := root.Exec(cleanup, "DROP SCHEMA "+name+" CASCADE")
		root.Close()
		if e != nil {
			t.Error(e)
		}
	})
	if e = store.Migrate(ctx, pool); e != nil {
		t.Fatal(e)
	}
	return pool
}
