package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	webserver "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/coreos/go-oidc"
	"github.com/honeycombio/beeline-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
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

type AWSRoleEnvironment struct {
	READER_ROLE_NAME string   `required:"true"`
	ORG_ROLE_ARNS    []string `required:"true"`
}

var concurrency int
var awsSessionRetries int
var skipAccountList []string

func init() {
	rootCmd.AddCommand(serveConfigCmd)
	serveConfigCmd.Flags().IntVar(&webServerPort, "web-server-port", 8080, "Port to host the aws config website")
	serveConfigCmd.Flags().IntVar(&concurrency, "concurrency", 1, "Number of parallel goroutines for account processing")
	serveConfigCmd.Flags().IntVar(&awsSessionRetries, "aws-retries", 5, "Number of times an AWS svc retries an operation")
	serveConfigCmd.Flags().StringSliceVar(&skipAccountList, "skip-accts", []string{}, "List of AWS account IDs serve-config should ignore.")
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

func loadAWSEnv() (*AWSRoleEnvironment, error) {
	env := &AWSRoleEnvironment{}
	err := envconfig.Process("AWS", env)
	if err != nil {
		return env, fmt.Errorf("Unable to load all the aws environment variables: %w", err)
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
	ctx, span := beeline.StartSpan(cmd.Context(), "serve-config run")
	defer span.Send()

	if concurrency == 0 {
		return errors.New("concurrency Limit cannot be 0")
	}

	// Initialize everything else
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
		return fmt.Errorf("Unable to create OIDC provider: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: oktaEnv.CLIENT_ID})

	oktaAppClient, err := createOktaClientApps(ctx, oktaEnv.ISSUER_URL, oktaEnv.PRIVATE_KEY, oktaEnv.SERVICE_CLIENT_ID)
	if err != nil {
		return fmt.Errorf("failed to create okta apps: %w", err)
	}

	configGenerationParams := webserver.AWSConfigGenerationParams{
		OIDCProvider:  oktaEnv.ISSUER_URL,
		AWSWorkerRole: awsEnv.READER_ROLE_NAME,
		AWSOrgRoles:   awsEnv.ORG_ROLE_ARNS,
		Concurrency:   concurrency,
		SkipAccounts:  sets.StringSet{},
	}

	configGenerationParams.SkipAccounts.Add(skipAccountList...)

	b, err := os.ReadFile("/rolemap/rolemap.yaml")
	if err != nil {
		return fmt.Errorf("reading rolemap.yaml: %w", err)
	}
	roleMappings := okta.OIDCRoleMapping{}
	err = yaml.Unmarshal(b, &roleMappings)
	if err != nil {
		return fmt.Errorf("unmarshalling rolemap.yaml: %w", err)
	}

	routerConfig := &webserver.RouterConfig{
		Verifier:            verifier,
		AwsGenerationParams: &configGenerationParams,
		OktaAppClient:       oktaAppClient,
	}

	router := webserver.GetRouter(ctx, routerConfig)
	port := fmt.Sprintf(":%d", webServerPort)

	return http.ListenAndServe(port, router)
}
