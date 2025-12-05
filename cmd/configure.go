package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var configURL string
var printOnly bool
var defaultRegion string
var defaultRoleName string

func init() {
	// required flags
	configureCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	configureCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	configureCmd.Flags().StringVar(&configURL, "config-url", "", "The URL of the config generation site.")
	configureCmd.MarkFlagRequired("client-id")  // nolint:errcheck
	configureCmd.MarkFlagRequired("issuer-url") // nolint:errcheck
	configureCmd.MarkFlagRequired("config-url") // nolint:errcheck

	// optional flags
	configureCmd.Flags().BoolVar(
		&printOnly,
		"print-only",
		false,
		`Set this flag if you don't want aws-oidc to modify your ~/.aws/config directly.
		 You can then configure your ~/.aws/config with the output.`,
	)
	configureCmd.Flags().StringVar(&defaultRegion, "default-region", "", "Region to configure for all profiles")
	configureCmd.Flags().StringVar(&defaultRoleName, "default-role-name", "", "Default role to configure for all profiles")

	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "aws-oidc configure",
	Long:  "Configure helps you configure your aws config. Depends on a config generation service running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := cli.GetToken(
			cmd.Context(),
			clientID,
			issuerURL,
			client.SetSuccessMessage(successMessage),
		)
		if err != nil {
			return err
		}

		config, err := aws_config_client.RequestConfig(cmd.Context(), token, configURL)
		if err != nil {
			return err
		}

		survey := &aws_config_client.Survey{}
		completer := aws_config_client.NewCompleter(
			survey,
			config,
			defaultRegion,
			defaultRoleName,
		)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}
		awsConfigPath := filepath.Join(homeDir, ".aws", "config")

		// create .aws dir if not present
		awsConfigDirPath := filepath.Dir(awsConfigPath)
		err = os.MkdirAll(awsConfigDirPath, 0775)
		if err != nil {
			return fmt.Errorf("could not create dir %s: %w", awsConfigPath, err)
		}

		// LooseLoad ignores the aws config file if missing
		originalConfig, err := ini.LooseLoad(awsConfigPath)
		if err != nil {
			return fmt.Errorf("could not open aws config: %w", err)
		}

		// We allow users to print aws config directly to stdout if they want
		// instead of us directly trying to modify their aws config
		if printOnly {
			return completer.Complete(originalConfig, &aws_config_client.AWSConfigSTDOUTWriter{})
		}

		awsConfigWriter := aws_config_client.NewAWSConfigFileWriter(awsConfigPath)
		err = completer.Complete(originalConfig, awsConfigWriter)
		if err != nil {
			return err
		}
		return awsConfigWriter.Finalize()
	},
}
