package main

import (
	"fmt"
	"os"

	"mixm/backend/internal/app"
)

func main() {
	if err := app.Seed(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
