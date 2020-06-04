package cmd

import (
	"fmt"
	"os"

	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	execCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	execCmd.Flags().StringVar(&roleARN, "aws-role-arn", "", "ARN value of role to assume")
	execCmd.MarkFlagRequired("client-id")    // nolint:errcheck
	execCmd.MarkFlagRequired("issuer-url")   // nolint:errcheck
	execCmd.MarkFlagRequired("aws-role-arn") // nolint:errcheck
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

	token, err := oidc.GetToken(ctx, clientID, issuerURL)
	if err != nil {
		return errors.Wrap(err, "Unable to obtain token from clientID and issuerURL")
	}

	tokenOutput, err := getter.GetAWSAssumeIdentity(ctx, token, roleARN)
	if err != nil {
		return errors.Wrap(err, "Unable to extract right token output from AWS Assume Web identity")
	}

	awsEnvVars := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", string(*tokenOutput.Credentials.AccessKeyId)),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", string(*tokenOutput.Credentials.SecretAccessKey)),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", string(*tokenOutput.Credentials.SessionToken)),
	}

	// our calculated awsEnvVars take precedence
	envVars := append(awsEnvVars, os.Environ()...)

	return exec(ctx, command, commandArgs, envVars)
}
