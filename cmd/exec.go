package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	"github.com/spf13/cobra"
)

type AWSDefaultEnvironment struct {
	DEFAULT_OUTPUT string
	DEFAULT_REGION string
}

func init() {
	execCmd.Flags().StringVar(
		&flagProfileName,
		aws_config_client.FlagProfile,
		"",
		"AWS Profile to fetch credentials from. Can set AWS_PROFILE instead.")

	execCmd.Flags().DurationVar(
		&sessionDuration,
		"session-duration",
		time.Hour,
		`The duration, of the role session. "1h" means 1 hour.
		Must be between 1-12 hours and must be <= the target role's max session duration.`,
	)

	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "aws-oidc exec",
	Long:  `exec takes in the command after the -- and executes it as a shell command`,
	Args:  parseArgs,
	RunE:  execRun,
}

func parseArgs(cmd *cobra.Command, args []string) error {
	dashIndex := cmd.ArgsLenAtDash()
	if dashIndex == -1 {
		return fmt.Errorf("please separate services and command with '--'.")
	}
	return nil
}

func execRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	dashIndex := cmd.ArgsLenAtDash()
	command := args[dashIndex]
	commandArgs := args[dashIndex+1:]

	awsOIDCConfig, err := aws_config_client.FetchParamsFromAWSConfig(
		cmd,
		aws_config_client.DefaultAWSConfigPath)
	if err != nil {
		return err
	}

	token, err := execGetToken(ctx, awsOIDCConfig.ClientID, awsOIDCConfig.IssuerURL)
	if err != nil {
		return fmt.Errorf("getting oidc token: %w", err)
	}

	assumeRoleOutput, err := getter.GetAWSAssumeIdentity(
		ctx,
		token,
		awsOIDCConfig.RoleARN,
		sessionDuration)
	if err != nil {
		return fmt.Errorf("extracting token output from AWS Assume Web identity: %w", err)
	}

	envVars := append(
		getAWSEnvVars(assumeRoleOutput, awsOIDCConfig),
		os.Environ()...,
	)

	return exec(command, commandArgs, envVars)
}
