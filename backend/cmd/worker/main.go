package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"mixm-edu/backend/internal/app"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/queue"
)

func main() {
	if e := run(); e != nil {
		slog.Error("worker stopped", "error", e)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, db, cache, e := app.Open(ctx)
	if e != nil {
		return e
	}
	defer db.Close()
	defer cache.Close()
	k, e := cluster.New(c)
	if e != nil {
		return e
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); queue.Dispatch(ctx, c, db) }()
	for i := 0; i < c.WorkerConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := queue.Work(ctx, c, db, cache, k); e != nil && ctx.Err() == nil {
				slog.Error("worker error", "error", e)
				cancel()
			}
		}()
	}
	wg.Wait()
	return nil
}
