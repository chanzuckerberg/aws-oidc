package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli"
	"github.com/chanzuckerberg/go-misc/oidc/v4/cli/client"
	"github.com/coreos/go-oidc"
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
		return fmt.Errorf("converting current credentials to json output: %w", err)
	}
	fmt.Println(string(output))

	return nil
}

func assumeRole(
	ctx context.Context,
	awsOIDCConfig *aws_config_client.AWSOIDCConfiguration,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	token, err := getOIDCToken(ctx, awsOIDCConfig)
	if err != nil {
		return nil, err
	}

	return getter.GetAWSAssumeIdentity(
		ctx,
		token,
		awsOIDCConfig.RoleARN,
		sessionDuration,
	)
}

func getOIDCToken(
	ctx context.Context,
	awsOIDCConfig *aws_config_client.AWSOIDCConfiguration,
) (*client.Token, error) {
	var token *client.Token
	var err error
	if deviceCodeFlow {
		token, err = cli.GetDeviceGrantToken(ctx, awsOIDCConfig.ClientID, awsOIDCConfig.IssuerURL, []string{
			oidc.ScopeOfflineAccess,
			oidc.ScopeOpenID,
			"profile",
			"groups",
		})
		if err != nil {
			return nil, fmt.Errorf("getting device grant token: %w", err)
		}
	} else {
		token, err = cli.GetToken(ctx, awsOIDCConfig.ClientID, awsOIDCConfig.IssuerURL)
		if err != nil {
			return nil, fmt.Errorf("getting authorization grant token: %w", err)
		}
	}
	return token, nil
}
