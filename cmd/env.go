package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/honeycombio/beeline-go"
	"github.com/spf13/cobra"
)

var flagProfileName string
var sessionDuration time.Duration

func init() {
	envCmd.Flags().StringVar(
		&flagProfileName,
		aws_config_client.FlagProfile,
		"",
		"AWS Profile to fetch credentials from.")

	envCmd.Flags().DurationVar(
		&sessionDuration,
		"session-duration",
		time.Hour,
		`The duration, of the role session. "1h" means 1 hour.
		Must be between 1-12 hours and must be <= the target role's max session duration.`,
	)

	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "aws-oidc env",
	Long: `Env will output relevant AWS credentials to environment variables.
	Useful when running docker such "docker run -it --env-file <(aws-oidc env --profile foobar) amazon/aws-cli sts get-caller-identity"
	`,
	RunE: envRun,
}

func envRun(cmd *cobra.Command, args []string) error {
	ctx, span := beeline.StartSpan(cmd.Context(), "env_command")
	defer span.Send()

	awsOIDCConfig, err := aws_config_client.FetchParamsFromAWSConfig(
		cmd,
		aws_config_client.DefaultAWSConfigPath)
	if err != nil {
		return err
	}

	assumeRoleOutput, err := assumeRole(ctx, awsOIDCConfig, sessionDuration)
	if err != nil {
		return err
	}

	// output in the appropriate format for docker
	envVars := getAWSEnvVars(assumeRoleOutput, awsOIDCConfig)
	fmt.Fprintln(os.Stdout, strings.Join(envVars, "\n"))
	return nil
}
