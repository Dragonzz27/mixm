package app

import (
	"context"
	"log/slog"

	"os/signal"
	"sync"
	"syscall"

	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/queue"
)

func RunWorker() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	c, db, cache, err := Open(ctx)
	if err != nil {
		return err
	}
	defer db.Close()
	defer cache.Close()
	k, err := cluster.New(c)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); queue.Dispatch(ctx, c, db) }()
	for i := 0; i < c.WorkerConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := queue.Work(ctx, c, db, cache, k); err != nil && ctx.Err() == nil {
				slog.Error("worker error", "error", err)
				cancel()
			}
		}()
	}
	wg.Wait()
	return nil
}
