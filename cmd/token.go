package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
	"github.com/coreos/go-oidc"

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

var tokenCmd = &cobra.Command{
	Use:           "token",
	Short:         "token prints the oidc tokens to stdout in json format",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdoutToken := &stdoutToken{
			Version: stdoutTokenVersion,
		}

		ctx := cmd.Context()

		token, err := execGetToken(ctx, clientID, issuerURL)
		if err != nil {
			return fmt.Errorf("getting oidc token: %w", err)
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

func execGetToken(ctx context.Context, clientID, issuerURL string) (*client.Token, error) {
	var clientOpts []client.OIDCClientOption
	if deviceCodeFlow {
		authenticator := client.NewDeviceGrantAuthenticator()
		clientOpts = append(clientOpts,
			client.WithDeviceGrantAuthenticator(authenticator),
			client.WithScopes([]string{
				oidc.ScopeOfflineAccess,
				oidc.ScopeOpenID,
				"profile",
				"email",
			}),
		)
	} else {
		clientOpts = append(clientOpts,
			client.WithAuthzGrantAuthenticator(
				client.DefaultAuthorizationGrantConfig,
				client.WithSuccessMessage(successMessage),
			),
			client.WithScopes(client.DefaultScopes),
		)
	}

	// For device code flow, redirect stderr to /dev/tty so auth prompts
	// are visible while keeping stderr available for error messages
	var restoreStderr func() = func() {}
	if deviceCodeFlow {
		restoreStderr = redirectStderrToTTY()
	}
	defer restoreStderr()

	token, err := cli.GetToken(
		ctx,
		clientID,
		issuerURL,
		cli.WithClientOptions(clientOpts...),
		cli.WithLocalCacheDir(nodeLocalCache),
	)

	if err != nil {
		return nil, fmt.Errorf("obtaining token from clientID and issuerURL: %w", err)
	}
	return token, nil
}

// redirectStderrToTTY redirects stderr to /dev/tty for interactive prompts
// and returns a function to restore the original stderr
func redirectStderrToTTY() func() {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		// If we can't open /dev/tty, just return a no-op
		return func() {}
	}

	originalStderr := os.Stderr
	os.Stderr = tty

	return func() {
		os.Stderr = originalStderr
		tty.Close()
	}
}
