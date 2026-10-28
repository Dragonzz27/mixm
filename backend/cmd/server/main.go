package main

import (
	"log/slog"
	"os"

	"mixm/backend/internal/app"
)

func main() {
	if err := app.RunServer(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
