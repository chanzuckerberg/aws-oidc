package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
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

	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "aws-oidc configure",
	Long:  "Configure helps you configure your aws config. Depends on a config generation service running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		survey := &aws_config_client.Survey{}
		completer := aws_config_client.NewCompleter(
			survey,
			generateDummyData(),
			issuerURL,
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

// For now generate dummy data, will later on use this for tests instead
// TODO(el): get rid of this in the next pr
func generateDummyData() map[server.ClientID][]server.ConfigProfile {
	configProfile1 := []server.ConfigProfile{
		{
			AcctName: "test1",
			RoleARN: arn.ARN{
				AccountID: "test_id_1",
				Resource:  "test1RoleName",
			},
		},
		{
			AcctName: "test2",
			RoleARN: arn.ARN{
				AccountID: "test_id_2",
				Resource:  "test2RoleName",
			},
		},
	}
	configProfile2 := []server.ConfigProfile{
		{
			AcctName: "foo1",
			RoleARN: arn.ARN{
				AccountID: "foo_id_1",
				Resource:  "foo1RoleName",
			},
		},
		{
			AcctName: "foo2",
			RoleARN: arn.ARN{
				AccountID: "foo_id_2",
				Resource:  "foo2RoleName",
			},
		},
	}

	data := map[server.ClientID][]server.ConfigProfile{}
	data["test_client_id"] = configProfile1
	data["foo_client_id"] = configProfile2
	return data
}
