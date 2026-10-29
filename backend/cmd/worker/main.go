package main

import (
	"log/slog"
	"os"

	"mixm/backend/internal/app"
)

// main 启动实例创建任务消费者。
func main() {
	if err := app.RunWorker(); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
