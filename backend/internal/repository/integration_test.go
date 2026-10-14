package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
	"mixm-edu/backend/internal/testutil"
)

func TestDefaultsNullZeroAndTransactions(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	if err := repository.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	row, err := repository.Write(ctx, db, "categories", "", model.Record{"name": "initial"})
	if err != nil {
		t.Fatal(err)
	}
	if !model.Bool(row, "is_active") || model.Number(row, "level") != 1 {
		t.Fatalf("defaults lost: %v", row)
	}
	row, err = repository.Write(ctx, db, "categories", model.String(row, "id"), model.Record{"is_active": false, "sort_order": 0, "description": nil})
	if err != nil || model.Bool(row, "is_active") || row["description"] != nil {
		t.Fatal(row, err)
	}
	sentinel := errors.New("rollback")
	err = db.Transaction(ctx, func(tx *repository.DB) error {
		_, err := repository.Write(ctx, tx, "categories", "", model.Record{"name": "rollback"})
		if err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	rows, err := repository.Query(ctx, db, "SELECT * FROM categories WHERE name=$1", "rollback")
	if err != nil || len(rows) != 0 {
		t.Fatal(rows, err)
	}
	role, err := repository.Write(ctx, db, "roles", "", model.Record{"name": "test", "route_keys": []string{"Users"}})
	if err != nil {
		t.Fatal(err)
	}
	rows, err = repository.Query(ctx, db, "SELECT * FROM roles WHERE route_keys ? $1 AND id=ANY($2)", "Users", []string{model.String(role, "id")})
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	if err = repository.Delete(ctx, db, "categories", model.String(row, "id")); err != nil {
		t.Fatal(err)
	}
	if _, err = repository.Get(ctx, db, "categories", model.String(row, "id")); !errors.Is(err, repository.ErrNotFound) {
		t.Fatal(err)
	}
}
func TestTaskLockPinsAndReleasesConnection(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	entered, release := make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := db.WithTaskLock(ctx, "task", func(conn *repository.DB) error { close(entered); <-release; return nil }); err != nil {
			t.Error(err)
		}
	}()
	<-entered
	ran := false
	if err := db.WithTaskLock(ctx, "task", func(*repository.DB) error { ran = true; return nil }); err != nil || ran {
		t.Fatal("duplicate acquired lock", err)
	}
	close(release)
	wg.Wait()
	if err := db.WithTaskLock(ctx, "task", func(*repository.DB) error { ran = true; return nil }); err != nil || !ran {
		t.Fatal("lock was not released", err)
	}
}

func TestTaskLockCancellationAndBrokenConnection(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	cancelled, cancel := context.WithCancel(ctx)
	err := db.WithTaskLock(cancelled, "cancel", func(conn *repository.DB) error { cancel(); return cancelled.Err() })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ran := false
	if err := db.WithTaskLock(ctx, "cancel", func(*repository.DB) error { ran = true; return nil }); err != nil || !ran {
		t.Fatal("cancel left lock held", err)
	}
	if err := db.WithTaskLock(ctx, "broken", func(conn *repository.DB) error {
		row, err := repository.One(ctx, conn, "SELECT pg_backend_pid() AS pid")
		if err != nil {
			return err
		}
		_, err = db.Exec(ctx, "SELECT pg_terminate_backend($1)", row["pid"])
		return err
	}); err != nil {
		t.Fatal(err)
	}
	ran = false
	if err := db.WithTaskLock(ctx, "broken", func(*repository.DB) error { ran = true; return nil }); err != nil || !ran {
		t.Fatal("broken connection returned to pool", err)
	}
}
