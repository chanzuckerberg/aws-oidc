package cmd

import (
	"context"
	"fmt"
	"net/http"

	webserver "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/coreos/go-oidc"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
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
	serveConfigCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to read the rolemap from")
	serveConfigCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap that holds the rolemap YAML")
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

func createOktaClientApps(ctx context.Context, orgURL, privateKey, oktaClientID string) (okta.AppLister, error) {
	oktaConfig := &okta.OktaClientConfig{
		ClientID:      oktaClientID,
		PrivateKeyPEM: privateKey,
		OrgURL:        orgURL,
	}
	client, err := okta.NewOktaClient(ctx, oktaConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Okta Client: %w", err)
	}
	return client, nil
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

	configMapName, err := cmd.Flags().GetString(flagConfigMapName)
	if err != nil {
		return fmt.Errorf("missing configmap-name flag: %w", err)
	}
	configMapKey, err := cmd.Flags().GetString(flagConfigMapKey)
	if err != nil {
		return fmt.Errorf("missing configmap-key flag: %w", err)
	}

	kubeClient, namespace, err := configmap.NewInClusterClient()
	if err != nil {
		return fmt.Errorf("creating in-cluster client: %w", err)
	}

	// Read the rolemap ConfigMap fresh on every request so the cronjob's updates are
	// served without a restart.
	mappingsProvider := func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error) {
		mappings, err := configmap.ReadRoleMappings(ctx, kubeClient, namespace, configMapName, configMapKey)
		if err != nil {
			return nil, err
		}
		return mappings.ByClientID(), nil
	}

	routerConfig := &webserver.RouterConfig{
		Verifier:            verifier,
		AwsGenerationParams: &configGenerationParams,
		OktaAppClient:       oktaAppClient,
		MappingsProvider:    mappingsProvider,
	}

	router := webserver.GetRouter(ctx, routerConfig)
	port := fmt.Sprintf(":%d", webServerPort)

	return http.ListenAndServe(port, router)
}
