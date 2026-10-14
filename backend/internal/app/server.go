package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"mixm-edu/backend/internal/infra/cluster"
	"mixm-edu/backend/internal/infra/queue"
	"mixm-edu/backend/internal/infra/ssh"
	"mixm-edu/backend/internal/router"
)

func RunServer() error {
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
		slog.Warn("Kubernetes unavailable; instance APIs return 503", "error", err)
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
			if err := (&ssh.Gateway{Cluster: k, Config: c}).ServeSSH(ctx, db); err != nil {
				slog.Error("SSH gateway stopped", "error", err)
				cancel()
			}
		}()
	}
	server := &http.Server{Addr: c.Address, Handler: router.New(c, db, cache, k), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	errs := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "address", c.Address, "go", "1.25.0")
		errs <- server.ListenAndServe()
	}()
	select {
	case err = <-errs:
		cancel()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, stop := context.WithTimeout(context.Background(), 15*time.Second)
		defer stop()
		err = server.Shutdown(shutdown)
		if err != nil {
			server.Close()
		}
		return err
	}
}
