package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v4/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/cobra"
)

func init() {
	tokenCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	tokenCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
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

// Get provider claims to discover endpoints
var providerClaims struct {
	DeviceAuthEndpoint  string   `json:"device_authorization_endpoint"`
	TokenEndpoint       string   `json:"token_endpoint"`
	GrantTypesSupported []string `json:"grant_types_supported"`
}

func useDeviceCodeFlow() bool {
	fmt.Fprintln(os.Stderr, "Device code flow is supported by this provider.")
	fmt.Fprintf(os.Stderr, "Use device code flow? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		return true
	}

	return false
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
		provider, err := oidc.NewProvider(ctx, issuerURL)
		if err != nil {
			return fmt.Errorf("creating oidc provider: %w", err)
		}

		if err := provider.Claims(&providerClaims); err != nil {
			return fmt.Errorf("getting provider claims: %w", err)
		}

		var token *client.Token
		if slices.Contains(providerClaims.GrantTypesSupported, "urn:ietf:params:oauth:grant-type:device_code") && useDeviceCodeFlow() {
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
