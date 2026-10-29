package main

import (
	"fmt"
	"os"

	"mixm/backend/internal/app"
)

// main 执行管理员创建命令。
func main() {
	if err := app.CreateAdmin(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
