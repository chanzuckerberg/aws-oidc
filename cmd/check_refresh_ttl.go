package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli"
	"github.com/spf13/cobra"
)

func init() {
	checkRefreshTTLCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	checkRefreshTTLCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	checkRefreshTTLCmd.MarkFlagRequired("client-id")  //nolint:errcheck
	checkRefreshTTLCmd.MarkFlagRequired("issuer-url") //nolint:errcheck

	rootCmd.AddCommand(checkRefreshTTLCmd)
}

func getNodeLocalCachePath(provided string) string {
	hostname, err := os.Hostname()
	if err != nil {
		return filepath.Join(provided, fmt.Sprintf("%d", os.Getuid()))
	}
	return filepath.Join(provided, hostname, fmt.Sprintf("%d", os.Getuid()))
}

var checkRefreshTTLCmd = &cobra.Command{
	Use:           "check-refresh-ttl",
	Short:         "Print the remaining TTL of the cached refresh token",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		path := getNodeLocalCachePath(nodeLocalCache)
		ttl, err := cli.CheckRefreshTokenTTL(
			ctx,
			clientID,
			issuerURL,
			cli.WithLocalCacheDir(path),
		)
		if err != nil {
			return fmt.Errorf("checking refresh token TTL: %w", err)
		}

		fmt.Fprintln(os.Stdout, ttl.String())
		return nil
	},
}
