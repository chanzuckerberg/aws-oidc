package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chanzuckerberg/aws-oidc/internal/agentstore"
	"github.com/chanzuckerberg/aws-oidc/internal/portal"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var portalPort int

func init() {
	rootCmd.AddCommand(servePortalCmd)
	servePortalCmd.Flags().IntVar(&portalPort, "web-server-port", 8080, "Port to host the portal on")
	servePortalCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to read the rolemap from")
	servePortalCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap that holds the rolemap YAML")
}

var servePortalCmd = &cobra.Command{
	Use:           "serve-portal",
	Short:         "aws-oidc serve-portal",
	Long:          "Start the agent-registry portal: a minimal UI to register agents and grant each a subset of your AWS access.",
	SilenceErrors: true,
	RunE:          servePortalRun,
}

func servePortalRun(cmd *cobra.Command, args []string) error {
	oktaEnv, err := loadOktaEnv()
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	oktaAppClient, err := createOktaClientApps(ctx, oktaEnv.ISSUER_URL, oktaEnv.PRIVATE_KEY, oktaEnv.SERVICE_CLIENT_ID)
	if err != nil {
		return fmt.Errorf("creating okta client: %w", err)
	}

	rolemapName, err := cmd.Flags().GetString(flagConfigMapName)
	if err != nil {
		return fmt.Errorf("missing configmap-name flag: %w", err)
	}
	rolemapKey, err := cmd.Flags().GetString(flagConfigMapKey)
	if err != nil {
		return fmt.Errorf("missing configmap-key flag: %w", err)
	}

	restConfig, namespace, err := configmap.NewInClusterConfig()
	if err != nil {
		return fmt.Errorf("loading in-cluster config: %w", err)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("creating dynamic client: %w", err)
	}

	// Read the rolemap fresh on each request so entitlements reflect the latest mapping.
	mappingsProvider := configmap.NewMappingsProvider(kubeClient, namespace, rolemapName, rolemapKey)

	store := agentstore.New(dynamicClient, namespace)

	// The gateway OIDC proxy authenticates with its own OAuth client (the shared
	// argus-global-oidc app), so the forwarded access token's cid is that client, not the
	// portal's own OKTA_CLIENT_ID. Configure it separately.
	gatewayClientID := os.Getenv("PORTAL_OIDC_CLIENT_ID")
	identity, err := portal.NewIdentityResolver(ctx, oktaEnv.ISSUER_URL, gatewayClientID)
	if err != nil {
		return fmt.Errorf("creating identity resolver: %w", err)
	}

	srv, err := portal.NewServer(portal.Config{
		Apps:             oktaAppClient,
		MappingsProvider: mappingsProvider,
		Store:            store,
		Identity:         identity,
		BasePath:         os.Getenv("PORTAL_BASE_PATH"),
	})
	if err != nil {
		return fmt.Errorf("creating portal server: %w", err)
	}

	addr := fmt.Sprintf(":%d", portalPort)
	return http.ListenAndServe(addr, srv.Handler())
}
