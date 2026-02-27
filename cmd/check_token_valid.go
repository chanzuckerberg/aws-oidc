package cmd

import (
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli"
	"github.com/spf13/cobra"
)

func init() {
	checkTokenValidCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	checkTokenValidCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	checkTokenValidCmd.MarkFlagRequired("client-id")  //nolint:errcheck
	checkTokenValidCmd.MarkFlagRequired("issuer-url") //nolint:errcheck

	rootCmd.AddCommand(checkTokenValidCmd)
}

var checkTokenValidCmd = &cobra.Command{
	Use:           "check-token-valid",
	Short:         "Check whether the cached OIDC token is present and valid",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		path := getNodeLocalCachePath(nodeLocalCache)
		err := cli.CheckTokenIsValid(
			ctx,
			clientID,
			issuerURL,
			cli.WithLocalCacheDir(path),
		)
		if err != nil {
			fmt.Fprintln(os.Stdout, "invalid")
			return fmt.Errorf("token is not valid: %w", err)
		}

		fmt.Fprintln(os.Stdout, "valid")
		return nil
	},
}
