//go:build linux || darwin
// +build linux darwin

package cmd

import (
	"fmt"
	"os"
	osexec "os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/pkg/errors"
)

func exec(command string, args []string, env []string) error {
	argv0, err := osexec.LookPath(command)
	if err != nil {
		return fmt.Errorf("Error finding command: %w", err)
	}

	argv := make([]string, 0, 1+len(args))
	argv = append(argv, command)
	argv = append(argv, args...)

	// Only return if the execution fails.
	return errors.Wrap(syscall.Exec(argv0, argv, env), "error executing command")
}

func getAWSEnvVars(assumeRoleOutput *sts.AssumeRoleWithWebIdentityOutput, awsOIDCConfig *aws_config_client.AWSOIDCConfiguration) []string {

	// Load config profile values if those environment variables don't exist (lowest precedence)
	envVars := []string{}
	_, present := os.LookupEnv("AWS_DEFAULT_REGION")
	if !present && (awsOIDCConfig.Region != nil) {
		envVars = append(envVars, fmt.Sprintf("AWS_DEFAULT_REGION=%s", *awsOIDCConfig.Region))
	}
	_, present = os.LookupEnv("AWS_DEFAULT_OUTPUT")
	if !present && (awsOIDCConfig.Output != nil) {
		envVars = append(envVars, fmt.Sprintf("AWS_DEFAULT_OUTPUT=%s", *awsOIDCConfig.Output))
	}

	// Load assumeRoleOutput credentials
	envVars = append(envVars,
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", string(*assumeRoleOutput.Credentials.AccessKeyId)),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", string(*assumeRoleOutput.Credentials.SecretAccessKey)),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", string(*assumeRoleOutput.Credentials.SessionToken)),
	)

	return envVars
}
