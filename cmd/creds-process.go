package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
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

	assumeRoleOutput, err := assumeRole(
		ctx,
		&aws_config_client.AWSOIDCConfiguration{
			ClientID:  clientID,
			IssuerURL: issuerURL,
			RoleARN:   roleARN,
		},
		time.Hour, // default to 1 hour
	)
	if err != nil {
		return err
	}

	creds := credProcess{
		Version:         credProcessVersion,
		AccessKeyID:     string(*assumeRoleOutput.Credentials.AccessKeyId),
		SecretAccessKey: string(*assumeRoleOutput.Credentials.SecretAccessKey),
		SessionToken:    string(*assumeRoleOutput.Credentials.SessionToken),
		Expiration:      assumeRoleOutput.Credentials.Expiration.Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(creds, "", "	")
	if err != nil {
		return errors.Wrap(err, "Unable to convert current credentials to json output")
	}
	fmt.Println(string(output))

	return nil
}

func assumeRole(
	ctx context.Context,
	awsOIDCConfig *aws_config_client.AWSOIDCConfiguration,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	token, err := oidc.GetToken(
		ctx,
		awsOIDCConfig.ClientID,
		awsOIDCConfig.IssuerURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to obtain OIDC token")
	}
	assumeRoleOutput, err := getter.GetAWSAssumeIdentity(
		ctx,
		token,
		awsOIDCConfig.RoleARN)
	return assumeRoleOutput, errors.Wrap(err, "unable to assume role")
}
