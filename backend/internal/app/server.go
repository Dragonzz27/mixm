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

	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/infra/queue"
	"mixm/backend/internal/infra/ssh"
	"mixm/backend/internal/router"
)

// RunServer 启动 HTTP 服务、任务发布器和 SSH 网关，并统一处理退出。
func RunServer() error {
	// 统一接收退出信号，供后台任务和网络服务协同停止。
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	c, db, cache, err := Open(ctx)
	if err != nil {
		return err
	}

	defer db.Close()
	defer cache.Close()

	// 初始化集群客户端；未配置集群时仍保留普通业务接口。
	k, err := cluster.New(c)
	if err != nil {
		slog.Warn("Kubernetes unavailable; instance APIs return 503", "error", err)
		k = nil
	}

	// 将发布器与 SSH 网关纳入退出等待，避免进程提前结束。
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		queue.Dispatch(ctx, c, db)
	}()
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

	// 启动 HTTP 服务，并在退出时等待正在处理的请求。
	server := &http.Server{
		Addr:              c.Address,
		Handler:           router.New(c, db, cache, k),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
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
