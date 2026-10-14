package config

import "testing"

func TestPostgresOnly(t *testing.T) {
	t.Setenv("SECRET_KEY", "test")
	t.Setenv("DATABASE_URL", "sqlite:///app.db")
	if _, e := Load(); e == nil {
		t.Fatal("SQLite runtime accepted")
	}
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("WORKER_CONCURRENCY", "0")
	if _, e := Load(); e == nil {
		t.Fatal("zero worker concurrency accepted")
	}
	t.Setenv("WORKER_CONCURRENCY", "4")
	c, e := Load()
	if e != nil {
		t.Fatal(e)
	}
	if c.WorkerConcurrency != 4 {
		t.Fatal("unexpected worker concurrency")
	}
}
