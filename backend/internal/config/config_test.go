package config

import "testing"

func TestPostgresOnly(t *testing.T) {
	t.Setenv("SECRET_KEY", "test")
	t.Setenv("DATABASE_URL", "sqlite:///app.db")
	if _, err := Load(); err == nil {
		t.Fatal("SQLite runtime accepted")
	}
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("WORKER_CONCURRENCY", "0")
	if _, err := Load(); err == nil {
		t.Fatal("zero worker concurrency accepted")
	}
	t.Setenv("WORKER_CONCURRENCY", "4")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.WorkerConcurrency != 4 {
		t.Fatal("unexpected worker concurrency")
	}
}
