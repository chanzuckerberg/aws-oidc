package cmd

import (
	"fmt"
	"os"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli"
	"github.com/spf13/cobra"
)

const flagOutputDurationInHours = "output-duration-in-hours"

func init() {
	checkRefreshTTLCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	checkRefreshTTLCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	checkRefreshTTLCmd.Flags().Bool(flagOutputDurationInHours, false, "Output the TTL as a decimal number of hours (machine-readable)")
	checkRefreshTTLCmd.MarkFlagRequired("client-id")  //nolint:errcheck
	checkRefreshTTLCmd.MarkFlagRequired("issuer-url") //nolint:errcheck

	rootCmd.AddCommand(checkRefreshTTLCmd)
}

var checkRefreshTTLCmd = &cobra.Command{
	Use:           "check-refresh-ttl",
	Short:         "Print the remaining TTL of the cached refresh token",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		outputInHours, err := cmd.Flags().GetBool(flagOutputDurationInHours)
		if err != nil {
			return fmt.Errorf("reading %s flag: %w", flagOutputDurationInHours, err)
		}

		ttl, err := cli.CheckRefreshTokenTTL(
			ctx,
			clientID,
			issuerURL,
			cli.WithLocalCacheDir(nodeLocalCacheFullPath),
		)
		if err != nil {
			return fmt.Errorf("checking refresh token TTL: %w", err)
		}

		if outputInHours {
			fmt.Fprintf(os.Stdout, "%.2f\n", ttl.Hours())
		} else {
			fmt.Fprintln(os.Stdout, ttl.String())
		}
		return nil
	},
}
