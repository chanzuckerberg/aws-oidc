package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	webserver "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	CZIOkta "github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/coreos/go-oidc"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var webServerPort int

type OktaWebserverEnvironment struct {
	PRIVATE_KEY       string
	SERVICE_CLIENT_ID string
	CLIENT_ID         string
	ISSUER_URL        string
}

type AWSEnvironment struct {
	READER_ROLE_NAME string
	MASTER_ROLE_ARNS []string
}

func init() {
	rootCmd.AddCommand(serveConfigCmd)
	serveConfigCmd.Flags().IntVar(&webServerPort, "web-server-port", 8080, "port to host the aws config website")
}

var serveConfigCmd = &cobra.Command{
	Use:           "serve-config",
	Short:         "aws-oidc serve-config",
	Long:          "Start a go webserver for returning client's aws config",
	SilenceErrors: true,
	RunE:          serveConfigRun,
}

func loadOktaEnv() (*OktaWebserverEnvironment, error) {
	env := &OktaWebserverEnvironment{}
	err := envconfig.Process("OKTA", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the environment variables")
	}
	return env, nil
}

func loadAWSEnv() (*AWSEnvironment, error) {
	env := &AWSEnvironment{}
	err := envconfig.Process("AWS", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the environment variables")
	}
	return env, nil
}

func createOktaClientApps(ctx context.Context, orgURL, privateKey, oktaClientID string) (okta.AppResource, error) {
	oktaConfig := &CZIOkta.OktaClientConfig{
		ClientID:      oktaClientID,
		PrivateKeyPEM: privateKey,
		OrgURL:        orgURL,
	}
	client, err := CZIOkta.NewOktaClient(ctx, oktaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create Okta Client")
	}
	return client.Application, nil
}

func serveConfigRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	oktaEnv, err := loadOktaEnv()
	if err != nil {
		return err
	}

	awsEnv, err := loadAWSEnv()
	if err != nil {
		return err
	}

	provider, err := oidc.NewProvider(ctx, oktaEnv.ISSUER_URL)
	if err != nil {
		return errors.Wrap(err, "Unable to create OIDC provider")
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oktaEnv.CLIENT_ID})

	awsSession, err := session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}
	)
	if err != nil {
		return errors.Wrap(err, "failed to create aws session")
	}

	oktaAppClient, err := createOktaClientApps(ctx, oktaEnv.ISSUER_URL, oktaEnv.PRIVATE_KEY, oktaEnv.SERVICE_CLIENT_ID)
	if err != nil {
		return errors.Wrap(err, "failed to create okta apps")
	}

	configGenerationParams := webserver.AWSConfigGenerationParams{
		OIDCProvider:   oktaEnv.ISSUER_URL,
		AWSWorkerRole:  awsEnv.READER_ROLE_NAME,
		AWSMasterRoles: awsEnv.MASTER_ROLE_ARNS,
	}

	getClientIDToProfiles, err := webserver.NewCachedGetClientIDToProfiles(
		ctx,
		&configGenerationParams,
		awsSession,
	)
	if err != nil {
		return errors.Wrap(err, "could not generate client id to aws role mapping")
	}

	routerConfig := &webserver.RouterConfig{
		Verifier:              verifier,
		AwsGenerationParams:   &configGenerationParams,
		OktaAppClient:         oktaAppClient,
		GetClientIDToProfiles: getClientIDToProfiles,
	}

	router := webserver.GetRouter(ctx, routerConfig)
	port := fmt.Sprintf(":%d", webServerPort)

	return http.ListenAndServe(port, router)
}
