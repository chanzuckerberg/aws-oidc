package cmd

// See https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html

import (
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/spf13/cobra"
)

func init() {
	browserCmd.Flags().StringVar(
		&flagProfileName,
		aws_config_client.FlagProfile,
		"",
		"AWS Profile to fetch credentials from. Can set AWS_PROFILE instead.")

	browserCmd.Flags().DurationVar(
		&sessionDuration,
		"session-duration",
		time.Hour,
		`The duration, of the role session. "1h" means 1 hour.
		Must be between 1-12 hours and must be <= the target role's max session duration.`,
	)

	rootCmd.AddCommand(browserCmd)
}

var browserCmd = &cobra.Command{
	Use:   "browser",
	Short: "aws-oidc browser",
	Long:  `browser opens a browser window and logs you in to the aws console.`,
	Args:  parseArgs,
	RunE:  execRun,
}

func browserRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

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

}
