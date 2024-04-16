package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl"
	"github.com/sirupsen/logrus"

	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var scopes []string

func init() {
	tokenCmd.Flags().StringVar(&clientID, "client-id", "", "client_id generated from the OIDC application")
	tokenCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider (either default IDP url or authorization server URL)")
	tokenCmd.MarkFlagRequired("client-id")  // nolint:errcheck
	tokenCmd.MarkFlagRequired("issuer-url") // nolint:errcheck
	tokenCmd.Flags().StringArrayVar(&scopes, "add-scope", []string{"openid", "offline_access", "email", "groups"}, "Individual scopes you want to add when fetching a token. separate them with a comma like this: --add-scope openid --add-scope email")
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
		clientOptions := []client.Option{}
		for _, scope := range scopes {
			logrus.Debugf("scope: %s", scope)
			clientOptions = append(clientOptions, client.AddScope(scope))
		}

		token, err := oidc_impl.GetToken(
			cmd.Context(),
			clientID,
			issuerURL,
			clientOptions...,
		)
		if err != nil {
			return err
		}

		stdoutToken.AccessToken = token.AccessToken
		stdoutToken.IDToken = token.IDToken
		stdoutToken.Expiry = token.Expiry

		data, err := json.Marshal(stdoutToken)
		if err != nil {
			return errors.Wrap(err, "could not json marshal oidc token")
		}

		_, err = fmt.Fprintln(os.Stdout, string(data))
		return errors.Wrap(err, "could not print token to stdout")
	},
}
