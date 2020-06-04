package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const credProcessVersion = 1

func init() {
	credProcessCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	credProcessCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	credProcessCmd.Flags().StringVar(&roleARN, "aws-role-arn", "", "ARN value of role to assume")
	credProcessCmd.MarkFlagRequired("client-id")    // nolint:errcheck
	credProcessCmd.MarkFlagRequired("issuer-url")   // nolint:errcheck
	credProcessCmd.MarkFlagRequired("aws-role-arn") // nolint:errcheck

	rootCmd.AddCommand(credProcessCmd)
}

type credProcess struct {
	Version         int    `json:"Version"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

// credProcessCmd represents the cred-process command
var credProcessCmd = &cobra.Command{
	Use:   "creds-process",
	Short: "aws-oidc creds-process",
	Long: `creds-process generates a credential_process ready output.
	--client-id, --issuerURL, and --aws-role-arn flags are required`,
	RunE: credProcessRun,
}

func credProcessRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	token, err := oidc.GetToken(ctx, clientID, issuerURL)
	if err != nil {
		return errors.Wrap(err, "Unable to obtain token from clientID and issuerURL")
	}

	tokenOutput, err := getter.GetAWSAssumeIdentity(ctx, token, roleARN)
	if err != nil {
		return errors.Wrap(err, "Unable to extract right token output from AWS Assume Web identity")
	}

	creds := credProcess{
		Version:         credProcessVersion,
		AccessKeyID:     string(*tokenOutput.Credentials.AccessKeyId),
		SecretAccessKey: string(*tokenOutput.Credentials.SecretAccessKey),
		SessionToken:    string(*tokenOutput.Credentials.SessionToken),
		Expiration:      token.Expiry.Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(creds, "", "	")
	if err != nil {
		return errors.Wrap(err, "Unable to convert current credentials to json output")
	}
	fmt.Println(string(output))

	return nil
}
