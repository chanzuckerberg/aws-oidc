package cmd

import (
	"os"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var configURL string

func init() {
	configureCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	configureCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	configureCmd.Flags().StringVar(&configURL, "config-url", "", "The URL of the config generation site.")
	configureCmd.MarkFlagRequired("client-id")  // nolint:errcheck
	configureCmd.MarkFlagRequired("issuer-url") // nolint:errcheck
	configureCmd.MarkFlagRequired("config-url") // nolint:errcheck

	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "aws-oidc configure",
	Long:  "Configure helps you configure your aws config. Depends on a config generation service running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := oidc.GetToken(cmd.Context(), clientID, issuerURL)
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
		)

		// TODO(el): should this be configurable?
		awsConfigPath, err := homedir.Expand("~/.aws/config")
		if err != nil {
			return errors.Wrap(err, "Could not parse aws config file path")
		}
		iniOut, err := ini.Load(awsConfigPath)
		if err != nil {
			return errors.Wrap(err, "could not open aws config")
		}

		err = completer.Loop(iniOut)
		if err != nil {
			return err
		}

		awsConfigFile, err := os.OpenFile(awsConfigPath, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer awsConfigFile.Close()

		_, err = iniOut.WriteTo(awsConfigFile)
		return errors.Wrap(err, "Could not write new aws config")
	},
}
