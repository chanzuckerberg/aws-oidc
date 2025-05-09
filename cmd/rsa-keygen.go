package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl"
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
		pub, err := oidc_impl.GenerateRSAKey()
		if err != nil {
			return fmt.Errorf("Unable to generate RSA key: %w", err)
		}

		b, err := pub.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}
