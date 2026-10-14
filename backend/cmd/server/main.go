package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"mixm-edu/backend/internal/api"
	"mixm-edu/backend/internal/app"
	"mixm-edu/backend/internal/cluster"
	"mixm-edu/backend/internal/queue"
)

func main() {
	if e := run(); e != nil {
		slog.Error("server stopped", "error", e)
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
		slog.Warn("Kubernetes unavailable; instance APIs return 503", "error", e)
		k = nil
	}
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	wg.Add(1)
	go func() { defer wg.Done(); queue.Dispatch(ctx, c, db) }()
	if k != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := k.ServeSSH(ctx, db); e != nil {
				slog.Error("SSH gateway stopped", "error", e)
				cancel()
			}
		}()
	}
	server := &http.Server{Addr: c.Address, Handler: api.New(c, db, cache, k), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	errs := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "address", c.Address, "go", "1.25.0")
		errs <- server.ListenAndServe()
	}()
	select {
	case e = <-errs:
		cancel()
		if errors.Is(e, http.ErrServerClosed) {
			return nil
		}
		return e
	case <-ctx.Done():
		shutdown, stop := context.WithTimeout(context.Background(), 15*time.Second)
		defer stop()
		e = server.Shutdown(shutdown)
		if e != nil {
			server.Close()
		}
		return e
	}
}
