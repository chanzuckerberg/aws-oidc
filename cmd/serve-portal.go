package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chanzuckerberg/aws-oidc/internal/portal"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/spf13/cobra"
)

var portalPort int

const (
	flagAgentConfigMapName = "agent-configmap-name"
	flagAgentConfigMapKey  = "agent-configmap-key"
)

func init() {
	rootCmd.AddCommand(servePortalCmd)
	servePortalCmd.Flags().IntVar(&portalPort, "web-server-port", 8080, "Port to host the portal on")
	servePortalCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to read the rolemap from")
	servePortalCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap that holds the rolemap YAML")
	servePortalCmd.Flags().String(flagAgentConfigMapName, "agent-registry", "Name of the ConfigMap that stores the agent registry")
	servePortalCmd.Flags().String(flagAgentConfigMapKey, "agents.yaml", "Key within the agent registry ConfigMap")
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
	agentCMName, err := cmd.Flags().GetString(flagAgentConfigMapName)
	if err != nil {
		return fmt.Errorf("missing agent-configmap-name flag: %w", err)
	}
	agentCMKey, err := cmd.Flags().GetString(flagAgentConfigMapKey)
	if err != nil {
		return fmt.Errorf("missing agent-configmap-key flag: %w", err)
	}

	kubeClient, namespace, err := configmap.NewInClusterClient()
	if err != nil {
		return fmt.Errorf("creating in-cluster client: %w", err)
	}

	// Read the rolemap fresh on each request so entitlements reflect the latest mapping.
	mappingsProvider := func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error) {
		mappings, err := configmap.ReadRoleMappings(ctx, kubeClient, namespace, rolemapName, rolemapKey)
		if err != nil {
			return nil, err
		}
		return mappings.ByClientID(), nil
	}

	store := portal.NewConfigMapStore(kubeClient, namespace, agentCMName, agentCMKey)

	srv, err := portal.NewServer(portal.Config{
		Apps:             oktaAppClient,
		MappingsProvider: mappingsProvider,
		Store:            store,
		Identity:         portal.NewIdentityResolver(),
	})
	if err != nil {
		return fmt.Errorf("creating portal server: %w", err)
	}

	addr := fmt.Sprintf(":%d", portalPort)
	return http.ListenAndServe(addr, srv.Handler())
}
