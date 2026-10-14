package main

import (
	"fmt"
	"os"

	"mixm-edu/backend/internal/app"
)

func main() {
	if err := app.CreateAdmin(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
