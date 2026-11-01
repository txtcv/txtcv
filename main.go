package main

import (
	"context"
	"fmt"
	"os"

	"github.com/txtcv/txtcv/commands"
)

func main() {
	cmd := commands.Root()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
