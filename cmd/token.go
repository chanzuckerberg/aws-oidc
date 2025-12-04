package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"

	"github.com/spf13/cobra"
)

var deviceCodeFlow bool

func init() {
	tokenCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	tokenCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	tokenCmd.Flags().BoolVar(&deviceCodeFlow, "device-code-flow", false, "Use device code flow for authentication")
	tokenCmd.MarkFlagRequired("client-id")  // nolint:errcheck
	tokenCmd.MarkFlagRequired("issuer-url") // nolint:errcheck

	rootCmd.AddCommand(tokenCmd)
}

const (
	stdoutTokenVersion = 1
)

type stdoutToken struct {
	Version int `json:"version,omitempty"`

	IDToken     string    `json:"id_token,omitempty"`
	AccessToken string    `json:"access_token,omitempty"`
	Expiry      time.Time `json:"expiry,omitempty"`
}

var tokenCmd = &cobra.Command{
	Use:           "token",
	Short:         "token prints the oidc tokens to stdout in json format",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdoutToken := &stdoutToken{
			Version: stdoutTokenVersion,
		}

		ctx := cmd.Context()

		var token *client.Token
		var err error
		if deviceCodeFlow {
			token, err = cli.GetDeviceGrantToken(ctx, clientID, issuerURL, []string{"openid", "profile", "offline_access"})
			if err != nil {
				return fmt.Errorf("getting device grant token: %w", err)
			}
		} else {
			token, err = cli.GetToken(ctx, clientID, issuerURL)
			if err != nil {
				return err
			}
		}

		stdoutToken.AccessToken = token.AccessToken
		stdoutToken.IDToken = token.IDToken
		stdoutToken.Expiry = token.Expiry

		data, err := json.Marshal(stdoutToken)
		if err != nil {
			return fmt.Errorf("could not json marshal oidc token: %w", err)
		}

		_, err = fmt.Fprintln(os.Stdout, string(data))
		if err != nil {
			return fmt.Errorf("could not print token to stdout: %w", err)
		}
		return nil
	},
}
