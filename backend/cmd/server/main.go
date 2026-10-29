package main

import (
	"log/slog"
	"os"

	"mixm/backend/internal/app"
)

// main 启动后端 API 与 SSH 网关。
func main() {
	if err := app.RunServer(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
