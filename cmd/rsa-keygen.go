package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(keygenCmd)
}

var keygenCmd = &cobra.Command{
	Use:           "rsa-keygen",
	Short:         "create a new rsa key for authenticating to Okta API",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		pub, err := cli.GenerateRSAKey()
		if err != nil {
			return fmt.Errorf("generating RSA key: %w", err)
		}

		b, err := pub.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}
