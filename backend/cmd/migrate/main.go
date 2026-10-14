package main

import (
	"context"
	"fmt"
	"os"

	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/store"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	c, e := config.Load()
	if e != nil {
		return e
	}
	ctx := context.Background()
	db, e := store.Open(ctx, c.DatabaseURL)
	if e != nil {
		return e
	}
	defer db.Close()
	return store.Migrate(ctx, db)
}
