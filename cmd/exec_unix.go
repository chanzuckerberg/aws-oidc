// +build linux darwin

package cmd

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/pkg/errors"
)

func exec(ctx context.Context, command string, args []string, env []string) error {
	argv0, err := osexec.LookPath(command)
	if err != nil {
		return errors.Wrap(err, "Error finding command")
	}

	argv := make([]string, 0, 1+len(args))
	argv = append(argv, command)
	argv = append(argv, args...)

	// Only return if the execution fails.
	return errors.Wrap(syscall.Exec(argv0, argv, env), "error executing command")
}

func getAWSEnvVars(assumeRoleOutput *sts.AssumeRoleWithWebIdentityOutput, awsOIDCConfig *aws_config_client.AWSOIDCConfiguration) []string {

	// Load config profile values (lowest precedence)
	envVars := []string{}
	_, present := os.LookupEnv("AWS_REGION")
	if (awsOIDCConfig.Region != "") && !present {
		envVars = append(envVars, fmt.Sprintf("AWS_REGION=%s", awsOIDCConfig.Region))
	}
	_, present = os.LookupEnv("AWS_OUTPUT")
	if awsOIDCConfig.Output != "" && !present {
		envVars = append(envVars, fmt.Sprintf("AWS_OUTPUT=%s", awsOIDCConfig.Output))
	}

	// Load assumeRoleOutput credentials
	envVars = append(envVars,
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", string(*assumeRoleOutput.Credentials.AccessKeyId)),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", string(*assumeRoleOutput.Credentials.SecretAccessKey)),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", string(*assumeRoleOutput.Credentials.SessionToken)),
	)

	// our environment variables take precedence
	envVars = append(envVars,
		os.Environ()...,
	)

	return envVars
}
