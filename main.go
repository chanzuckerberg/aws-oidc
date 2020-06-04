package main

import (
	"context"

	"github.com/chanzuckerberg/aws-oidc/cmd"
	"github.com/sirupsen/logrus"
)

func exec() error {
	return cmd.Execute(context.Background())
}

func main() {
	err := exec()
	if err != nil {
		logrus.Fatal(err)
	}
}
