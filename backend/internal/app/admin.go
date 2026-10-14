package app

import (
	"context"
	"fmt"
	"os"

	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
)

func CreateAdmin() error {
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_USERNAME")
	if name == "" {
		name = "admin"
	}
	if email == "" || password == "" {
		return fmt.Errorf("set ADMIN_EMAIL and ADMIN_PASSWORD")
	}
	c, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	db, err := repository.Open(ctx, c.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err = repository.Migrate(ctx, db); err != nil {
		return err
	}
	h, err := auth.Hash(password)
	if err != nil {
		return err
	}
	_, err = repository.Write(ctx, db, "users", "", model.Record{"username": name, "email": email, "hashed_password": h, "is_superuser": true})
	if err == nil {
		fmt.Println("管理员已创建")
	}
	return err
}
