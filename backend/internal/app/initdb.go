package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"mixm/backend/internal/config"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

// Seed 初始化默认角色、内容类型和开发示例，可重复执行。
func Seed() error {
	// 读取配置并应用编号迁移，保证初始化所需的表已存在。
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

	// 基础配置和示例内容在同一事务中初始化，失败时整体回滚。
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// 分别定义学生菜单和管理员菜单的权限范围。
	client := []string{
		"ClientHome",
		"ClientContents",
		"ClientMyInstances",
		"ClientAssignments",
		"ClientSimulationTest",
	}
	manager := append(
		append([]string{}, client...),
		"Dashboard",
		"Users",
		"Roles",
		"Categories",
		"Contents",
		"Templates",
		"Bindings",
		"Clusters",
		"Instances",
		"AdminImageRepositories",
		"Assignments",
		"Questions",
		"Practices",
	)

	// 按角色名称执行幂等写入，重复运行不会产生重复角色。
	for _, role := range []struct {
		name string
		keys []string
	}{
		{"学生", client},
		{"管理员", manager},
	} {
		if _, err = repository.SeedRole(ctx, tx, uuid.NewString(), role.name, model.JSON(role.keys)); err != nil {
			return err
		}
	}

	// 按稳定的内容类型代码初始化默认分类。
	for code, name := range map[string]string{
		"tutorial":      "教程",
		"documentation": "文档",
		"experiment":    "实验",
		"video":         "视频",
		"exercise":      "练习",
	} {
		if _, err = repository.SeedContentType(ctx, tx, uuid.NewString(), code, name); err != nil {
			return err
		}
	}

	if err = seedExamples(ctx, tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err == nil {
		fmt.Println("角色、内容类型和开发示例初始化完成：3 篇教学内容、3 份作业、5 道题目、1 份练习")
	}

	return err
}
