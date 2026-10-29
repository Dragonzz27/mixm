package app

import (
	"context"
	"fmt"
	"os"

	"mixm/backend/internal/auth"
	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

// CreateAdmin 读取管理员参数，初始化数据库结构并创建管理员账号。
func CreateAdmin() error {
	// 读取管理员参数，并在连接数据库前检查必填凭据。
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_USERNAME")
	if name == "" {
		name = "admin"
	}

	if email == "" || password == "" {
		return fmt.Errorf("set ADMIN_EMAIL and ADMIN_PASSWORD")
	}

	// 加载环境配置，准备管理员账号使用的数据库结构。
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

	// 仅保存密码哈希，随后写入超级管理员账号。
	h, err := auth.Hash(password)
	if err != nil {
		return err
	}

	_, err = repository.Write(
		ctx,
		db,
		"users",
		"",
		model.Record{
			"username":        name,
			"email":           email,
			"hashed_password": h,
			"is_superuser":    true,
		},
	)
	if err == nil {
		fmt.Println("管理员已创建")
	}

	return err
}
