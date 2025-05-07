package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	webserver "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/coreos/go-oidc"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert/yaml"
)

var webServerPort int

type OktaWebserverEnvironment struct {
	PRIVATE_KEY       string `required:"true"`
	SERVICE_CLIENT_ID string `required:"true"`
	CLIENT_ID         string `required:"true"`
	ISSUER_URL        string `required:"true"`
}

func init() {
	rootCmd.AddCommand(serveConfigCmd)
	serveConfigCmd.Flags().IntVar(&webServerPort, "web-server-port", 8080, "Port to host the aws config website")
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
		return env, fmt.Errorf("Unable to load all the okta environment variables: %w", err)
	}
	return env, nil
}

func createOktaClientApps(ctx context.Context, orgURL, privateKey, oktaClientID string) (okta.AppResource, error) {
	oktaConfig := &okta.OktaClientConfig{
		ClientID:      oktaClientID,
		PrivateKeyPEM: privateKey,
		OrgURL:        orgURL,
	}
	client, err := okta.NewOktaClient(ctx, oktaConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Okta Client: %w", err)
	}
	return client.Application, nil
}

func serveConfigRun(cmd *cobra.Command, args []string) error {
	oktaEnv, err := loadOktaEnv()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	provider, err := oidc.NewProvider(ctx, oktaEnv.ISSUER_URL)
	if err != nil {
		return fmt.Errorf("Unable to create OIDC provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oktaEnv.CLIENT_ID})

	oktaAppClient, err := createOktaClientApps(ctx, oktaEnv.ISSUER_URL, oktaEnv.PRIVATE_KEY, oktaEnv.SERVICE_CLIENT_ID)
	if err != nil {
		return fmt.Errorf("failed to create okta apps: %w", err)
	}

	configGenerationParams := webserver.AWSConfigGenerationParams{
		OIDCProvider: oktaEnv.ISSUER_URL,
	}

	b, err := os.ReadFile("/rolemap/rolemap.yaml")
	if err != nil {
		return fmt.Errorf("reading rolemap.yaml: %w", err)
	}
	clientMappings := okta.OIDCRoleMappings{}
	err = yaml.Unmarshal(b, &clientMappings)
	if err != nil {
		return fmt.Errorf("unmarshalling rolemap.yaml: %w", err)
	}
	clientMappingsByKey := make(okta.OIDCRoleMappingsByKey)
	for _, mapping := range clientMappings {
		_, ok := clientMappingsByKey[mapping.OktaClientID]
		if ok {
			clientMappingsByKey[mapping.OktaClientID] = append(clientMappingsByKey[mapping.OktaClientID], mapping)
		} else {
			clientMappingsByKey[mapping.OktaClientID] = []okta.OIDCRoleMapping{mapping}
		}
	}
	routerConfig := &webserver.RouterConfig{
		Verifier:            verifier,
		AwsGenerationParams: &configGenerationParams,
		OktaAppClient:       oktaAppClient,
		ClientMappings:      clientMappingsByKey,
	}

	router := webserver.GetRouter(ctx, routerConfig)
	port := fmt.Sprintf(":%d", webServerPort)

	return http.ListenAndServe(port, router)
}
