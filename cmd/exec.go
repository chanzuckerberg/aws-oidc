package cmd

import (
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	oidc_client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type AWSDefaultEnvironment struct {
	DEFAULT_OUTPUT string
	DEFAULT_REGION string
}

func loadAWSDefaultEnv() (*AWSDefaultEnvironment, error) {
	env := &AWSDefaultEnvironment{}
	err := envconfig.Process("AWS", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the aws environment variables")
	}
	return env, nil
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
		return errors.New("please separate services and command with '--'.")
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

	token, err := oidc.GetToken(
		ctx,
		awsOIDCConfig.ClientID,
		awsOIDCConfig.IssuerURL,
		oidc_client.SetSuccessMessage(successMessage),
	)
	if err != nil {
		return errors.Wrap(err, "Unable to obtain token from clientID and issuerURL")
	}

	assumeRoleOutput, err := getter.GetAWSAssumeIdentity(
		ctx,
		token,
		awsOIDCConfig.RoleARN,
		sessionDuration)
	if err != nil {
		return errors.Wrap(err, "Unable to extract right token output from AWS Assume Web identity")
	}

	envVars := getAWSEnvVars(assumeRoleOutput, awsOIDCConfig)

	return exec(ctx, command, commandArgs, envVars)
}
