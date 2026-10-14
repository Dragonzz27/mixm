package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
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
	if e = store.Migrate(ctx, db); e != nil {
		return e
	}
	tx, e := db.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	client := []string{"ClientHome", "ClientContents", "ClientMyInstances", "ClientAssignments", "ClientSimulationTest"}
	manager := append(append([]string{}, client...), "Dashboard", "Users", "Roles", "Categories", "Contents", "Templates", "Bindings", "Clusters", "Instances", "AdminImageRepositories", "Assignments", "Questions", "Practices")
	for _, role := range []struct {
		name string
		keys []string
	}{{"学生", client}, {"管理员", manager}} {
		if _, e = tx.Exec(ctx, "INSERT INTO roles(id,name,route_keys) VALUES($1,$2,$3) ON CONFLICT(name) DO NOTHING", uuid.NewString(), role.name, store.JSON(role.keys)); e != nil {
			return e
		}
	}
	for code, name := range map[string]string{"tutorial": "教程", "documentation": "文档", "experiment": "实验", "video": "视频", "exercise": "练习"} {
		if _, e = tx.Exec(ctx, "INSERT INTO content_types(id,code,name) VALUES($1,$2,$3) ON CONFLICT(code) DO NOTHING", uuid.NewString(), code, name); e != nil {
			return e
		}
	}
	if e = tx.Commit(ctx); e == nil {
		fmt.Println("角色和内容类型初始化完成")
	}
	return e
}
