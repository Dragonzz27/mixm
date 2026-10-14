package main

import (
	"context"
	"fmt"
	"os"

	"mixm-edu/backend/internal/auth"
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
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_USERNAME")
	if name == "" {
		name = "admin"
	}
	if email == "" || password == "" {
		return fmt.Errorf("set ADMIN_EMAIL and ADMIN_PASSWORD")
	}
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
	if e = store.Migrate(ctx, db); e != nil {
		return e
	}
	h, e := auth.Hash(password)
	if e != nil {
		return e
	}
	_, e = store.Write(ctx, db, "users", "", store.Row{"username": name, "email": email, "hashed_password": h, "is_superuser": true})
	if e == nil {
		fmt.Println("管理员已创建")
	}
	return e
}
