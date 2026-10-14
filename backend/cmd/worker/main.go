package main

import (
	"log/slog"
	"os"

	"mixm-edu/backend/internal/app"
)

func main() {
	if err := app.RunWorker(); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
