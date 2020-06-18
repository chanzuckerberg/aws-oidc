package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/spf13/cobra"
)

var flagProfileName string

func init() {
	envCmd.Flags().StringVar(
		&flagProfileName,
		aws_config_client.FlagProfile,
		"",
		"AWS Profile to fetch credentials from.")

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
	ctx := cmd.Context()
	awsOIDCConfig, err := aws_config_client.FetchParamsFromAWSConfig(
		cmd,
		aws_config_client.DefaultAWSConfigPath)
	if err != nil {
		return err
	}

	assumeRoleOutput, err := assumeRole(ctx, awsOIDCConfig)
	if err != nil {
		return err
	}

	// output in the appropriate format for docker
	envVars := getAWSEnvVars(assumeRoleOutput)
	fmt.Fprintln(os.Stdout, strings.Join(envVars, "\n"))
	return nil
}
