package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	execCmd.Flags().StringVar(
		&flagProfileName,
		aws_config_client.FlagProfile,
		"",
		"AWS Profile to fetch credentials from. Can set AWS_PROFILE instead.")

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

	token, err := oidc.GetToken(ctx, awsOIDCConfig.ClientID, awsOIDCConfig.IssuerURL)
	if err != nil {
		return errors.Wrap(err, "Unable to obtain token from clientID and issuerURL")
	}

	assumeRoleOutput, err := getter.GetAWSAssumeIdentity(ctx, token, awsOIDCConfig.RoleARN)
	if err != nil {
		return errors.Wrap(err, "Unable to extract right token output from AWS Assume Web identity")
	}

	// our calculated awsEnvVars take precedence
	envVars := append(
		getAWSEnvVars(assumeRoleOutput),
		os.Environ()...,
	)

	return exec(ctx, command, commandArgs, envVars)
}

func getAWSEnvVars(assumeRoleOutput *sts.AssumeRoleWithWebIdentityOutput) []string {
	return []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", string(*assumeRoleOutput.Credentials.AccessKeyId)),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", string(*assumeRoleOutput.Credentials.SecretAccessKey)),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", string(*assumeRoleOutput.Credentials.SessionToken)),
	}
}
