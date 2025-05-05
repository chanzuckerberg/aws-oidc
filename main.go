package main

import (
	"context"
	"log/slog"

	"github.com/chanzuckerberg/aws-oidc/cmd"
)

func exec() error {
	return cmd.Execute(context.Background())
}

func main() {
	err := exec()
	if err != nil {
		slog.Error("failed to execute command", "error", err)
	}
}
