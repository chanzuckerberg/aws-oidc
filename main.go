package main

import (
	"context"
	"os"

	"github.com/chanzuckerberg/aws-oidc/cmd"
)

func main() {
	err := cmd.Execute(context.Background())
	if err != nil {
		// exit code is needed to indicate to the AWS credential process that the command failed
		os.Exit(1)
	}
}
