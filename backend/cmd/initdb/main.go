package main

import (
	"fmt"
	"os"

	"mixm/backend/internal/app"
)

// main 执行基础数据初始化命令。
func main() {
	if err := app.Seed(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
