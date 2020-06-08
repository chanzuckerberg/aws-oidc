package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/spf13/cobra"
)

func init() {
	configureCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	configureCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	configureCmd.Flags().StringVar(&configURL, "config-url", "", "The URL of the config generation site.")

	rootCmd.AddCommand(configureCmd)
}

var configURL string

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "aws-oidc configure",
	Long:  "Configure helps you configure your aws config. Depends on a config generation service running.",
	RunE: func(cmd *cobra.Command, args []string) error {
		completer := aws_config_client.NewCompleter(generateDummyData())

		account := ""

		err := survey.AskOne(
			&survey.Select{
				Options: completer.CompleteAccount(),
			},
			&account,
		)

		return err
	},
}

func generateDummyData() map[server.ClientID][]server.ConfigProfile {
	configProfile1 := []server.ConfigProfile{
		{
			AcctName: "test1",
			RoleARN: arn.ARN{
				AccountID: "test_id_1",
			},
		},
		{
			AcctName: "test2",
			RoleARN: arn.ARN{
				AccountID: "test_id_2",
			},
		},
	}
	configProfile2 := []server.ConfigProfile{
		{
			AcctName: "foo1",
			RoleARN: arn.ARN{
				AccountID: "foo_id_1",
			},
		},
		{
			AcctName: "foo2",
			RoleARN: arn.ARN{
				AccountID: "foo_id_2",
			},
		},
	}

	data := map[server.ClientID][]server.ConfigProfile{}
	data["test_client_id"] = configProfile1
	data["foo_client_id"] = configProfile2
	return data
}
