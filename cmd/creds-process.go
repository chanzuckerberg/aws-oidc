package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chanzuckerberg/go-misc/osutil"
	"github.com/chanzuckerberg/go-misc/pidlock"
	"github.com/mitchellh/go-homedir" // for storage, refactor out.
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/getter"
	oidc "github.com/chanzuckerberg/go-misc/oidc_cli"
	"github.com/chanzuckerberg/go-misc/oidc_cli/storage" // Not sure why this is also necessary, given above line?
	oidc_client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	credProcessCmd.Flags().StringVar(&clientID, "client-id", "", "CLIENT_ID generated from the OIDC application")
	credProcessCmd.Flags().StringVar(&issuerURL, "issuer-url", "", "The URL that hosts the OIDC identity provider")
	credProcessCmd.Flags().StringVar(&roleARN, "aws-role-arn", "", "ARN value of role to assume")
	credProcessCmd.MarkFlagRequired("client-id")    // nolint:errcheck
	credProcessCmd.MarkFlagRequired("issuer-url")   // nolint:errcheck
	credProcessCmd.MarkFlagRequired("aws-role-arn") // nolint:errcheck

	rootCmd.AddCommand(credProcessCmd)
}

// credProcessCmd represents the cred-process command
var credProcessCmd = &cobra.Command{
	Use:   "creds-process",
	Short: "aws-oidc creds-process",
	Long: `creds-process generates a credential_process ready output.
	--client-id, --issuerURL, and --aws-role-arn flags are required`,
	RunE: credProcessRun,
}

const (
	lockFilePath          = "/tmp/aws-oidc-cred.lock"
	defaultFileStorageDir = "~/.oidc-cli" //Do we need a different storage directory from the rest of the CLI, here?
	assumeRoleTime = time.Hour // default to 1 hour

)


func updateCred(ctx context.Context,
	awsOIDCConfig *aws_config_client.AWSOIDCConfiguration) (*processedCred, error) {
		assumeRoleOutput, err := assumeRole(
			ctx,
			awsOIDCConfig,
			assumeRoleTime,
		)
		if err != nil {
			return nil, err //Is this correct error handling when err is second result?
		}
	
		creds := processedCred{
			Version:         processedCredVersion,
			AccessKeyID:     string(*assumeRoleOutput.Credentials.AccessKeyId),
			SecretAccessKey: string(*assumeRoleOutput.Credentials.SecretAccessKey),
			SessionToken:    string(*assumeRoleOutput.Credentials.SessionToken),
			Expiration:      assumeRoleOutput.Credentials.Expiration.Format(time.RFC3339),
			CacheExpiry: 	 *assumeRoleOutput.Credentials.Expiration,
		}
		return &creds, nil // correct for when no error?
	
		output, err := json.MarshalIndent(creds, "", "	")
		if err != nil {
			return nil, errors.Wrap(err, "Unable to convert current credentials to json output") //error handling? as above
		}
		fmt.Println(string(output))
	
		return nil, nil
	}

func credProcessRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	fileLock, err := pidlock.NewLock(lockFilePath)
	if err != nil {
		return errors.Wrap(err, "unable to create lock")
	}

	config := &aws_config_client.AWSOIDCConfiguration{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		RoleARN:   roleARN,
	}
	storage, err := getStorage(clientID, issuerURL)
	if err != nil {
		return err
	}

	cache := NewCache(storage, updateCred, fileLock)

	creds, err := cache.Read(ctx, config)
	if err != nil {
		return errors.Wrap(err, "Unable to process credentials.")
	}
	if creds == nil {
		return errors.New("nil token from OIDC-IDP")
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
	ctx, span := beeline.StartSpan(ctx, "assumeAWSRole")
	defer span.Send()

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
) (*oidc_client.Token, error) {
	ctx, span := beeline.StartSpan(ctx, "get_oidc_token")
	defer span.Send()

	return oidc.GetToken(
		ctx,
		awsOIDCConfig.ClientID,
		awsOIDCConfig.IssuerURL,
		oidc_client.SetSuccessMessage(successMessage),
	)
}

// COPYPASTA from https://github.com/chanzuckerberg/go-misc/blob/53701f2e26a17e5dee3e5ee793478f442679764c/oidc_cli/token_getter.go
// TODO - Refactor
func getStorage(clientID string, issuerURL string) (storage.Storage, error) {
	isWSL, err := osutil.IsWSL()
	if err != nil {
		return nil, err
	}

	// If WSL we use a file storage which does not cache refreshTokens
	//    we do this because WSL doesn't have a graphical interface
	//    and therefore limits how we can interact with a keyring (such as gnome-keyring).
	// To limit the risks of having a long-lived refresh token around,
	//    we disable this part of the flow for WSL. This could change in the future
	//    when we find a better way to work with a WSL secure storage.
	if isWSL {
		return getFileStorage(clientID, issuerURL)
	}

	return storage.NewKeyring(clientID, issuerURL), nil
}

func getFileStorage(clientID string, issuerURL string) (storage.Storage, error) {
	dir, err := homedir.Expand(defaultFileStorageDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not expand path")
	}

	return storage.NewFile(dir, clientID, issuerURL), nil
}