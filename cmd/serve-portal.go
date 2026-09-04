package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/chanzuckerberg/aws-oidc/internal/agentdefaults"
	"github.com/chanzuckerberg/aws-oidc/internal/agentstore"
	"github.com/chanzuckerberg/aws-oidc/internal/portal"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	flagAgentRuntime             = "agent-runtime"
	flagAgentTailscale           = "tailscale"
	flagAgentMaxCPU              = "agent-max-cpu"
	flagAgentMaxMemory           = "agent-max-memory"
	flagAgentDefaultImage        = "agent-default-image"
	flagAgentDefaultStorageClass = "agent-default-storage-class"
)

var portalPort int

func init() {
	rootCmd.AddCommand(servePortalCmd)
	servePortalCmd.Flags().IntVar(&portalPort, "web-server-port", 8080, "Port to host the portal on")
	servePortalCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to read the rolemap from")
	servePortalCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap that holds the rolemap YAML")
	servePortalCmd.Flags().Bool(flagAgentRuntime, false, "Offer running an agent's threads as pods in the cluster; set this only where the operator is configured to run them")
	servePortalCmd.Flags().Bool(flagAgentTailscale, false, "Show the Tailscale page in the agent sidebar; set this only where the operator is configured for tailnet enrollment")
	servePortalCmd.Flags().String(flagAgentMaxCPU, "4", "Most CPU an agent thread may request")
	servePortalCmd.Flags().String(flagAgentMaxMemory, "16Gi", "Most memory an agent thread may request")
	servePortalCmd.Flags().Int(flagMaxThreadsPerAgent, 5, "Maximum threads one agent may run")
	servePortalCmd.Flags().String(flagAgentDefaultImage, os.Getenv("AGENT_DEFAULT_IMAGE"), "Default container image shown in the agent form; blank means the form shows an empty placeholder")
	servePortalCmd.Flags().String(flagAgentDefaultStorageClass, os.Getenv("AGENT_DEFAULT_STORAGE_CLASS"), "Default storage class pre-filled in the agent form")
	servePortalCmd.Flags().String(flagDefaultsConfig, os.Getenv("AGENT_DEFAULTS_CONFIG"), "Path to the agent-defaults YAML file mounted from the agent-defaults ConfigMap; when set, overrides static default flags without a restart")
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
	agentRuntime, err := cmd.Flags().GetBool(flagAgentRuntime)
	if err != nil {
		return fmt.Errorf("missing agent-runtime flag: %w", err)
	}
	agentTailscale, err := cmd.Flags().GetBool(flagAgentTailscale)
	if err != nil {
		return fmt.Errorf("missing tailscale flag: %w", err)
	}
	limits, err := agentLimits(cmd)
	if err != nil {
		return err
	}
	defaultsConfigPath, err := cmd.Flags().GetString(flagDefaultsConfig)
	if err != nil {
		return fmt.Errorf("missing defaults-config flag: %w", err)
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

	githubApp, err := githubAppFromEnv()
	if err != nil {
		return fmt.Errorf("configuring github repositories: %w", err)
	}

	cfg := portal.Config{
		Apps:             oktaAppClient,
		MappingsProvider: mappingsProvider,
		Store:            store,
		Identity:         identity,
		BasePath:         os.Getenv("PORTAL_BASE_PATH"),
		AgentRuntime:     agentRuntime,
		AgentTailscale:   agentTailscale,
		Limits:           limits,
		Namespace:        namespace,
		DefaultsLoader:   agentdefaults.NewLoader(defaultsConfigPath),
	}
	// Assign only when configured: a nil *GitHubApp stored in the interface field would read
	// as non-nil and turn the Repositories page on without a working backend.
	if githubApp != nil {
		cfg.Repositories = githubApp
	}

	srv, err := portal.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("creating portal server: %w", err)
	}

	addr := fmt.Sprintf(":%d", portalPort)
	slog.Info("portal listening",
		"addr", addr,
		"base_path", os.Getenv("PORTAL_BASE_PATH"),
		"namespace", namespace,
		"rolemap_configmap", rolemapName,
		"agent_runtime", agentRuntime,
		"agent_tailscale", agentTailscale,
		"agent_repositories", githubApp != nil,
	)
	return http.ListenAndServe(addr, srv.Handler())
}

// githubAppFromEnv builds the GitHub App client the Repositories page uses to search and
// validate repositories. It reads the same credentials the operator uses: the app id, the
// default installation, the owner-to-installation map, and the private key (inline in
// GITHUB_APP_PRIVATE_KEY or a path in GITHUB_APP_PRIVATE_KEY_FILE). Returns nil when the app
// id or key is absent, which leaves the Repositories page hidden.
func githubAppFromEnv() (*portal.GitHubApp, error) {
	key, err := githubAppPrivateKey()
	if err != nil {
		return nil, err
	}
	return portal.NewGitHubApp(
		os.Getenv("GITHUB_APP_ID"),
		key,
		os.Getenv("GITHUB_APP_INSTALLATION_ID"),
		os.Getenv("GITHUB_APP_INSTALLATION_MAP"),
		os.Getenv("GITHUB_API_URL"),
	)
}

func githubAppPrivateKey() ([]byte, error) {
	if pem := os.Getenv("GITHUB_APP_PRIVATE_KEY"); pem != "" {
		return []byte(pem), nil
	}
	path := os.Getenv("GITHUB_APP_PRIVATE_KEY_FILE")
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading github app private key file: %w", err)
	}
	return data, nil
}

// agentLimits reads the ceilings an agent owner is held to when sizing their threads.
func agentLimits(cmd *cobra.Command) (portal.AgentLimits, error) {
	maxCPU, err := cmd.Flags().GetString(flagAgentMaxCPU)
	if err != nil {
		return portal.AgentLimits{}, fmt.Errorf("missing agent-max-cpu flag: %w", err)
	}
	maxMemory, err := cmd.Flags().GetString(flagAgentMaxMemory)
	if err != nil {
		return portal.AgentLimits{}, fmt.Errorf("missing agent-max-memory flag: %w", err)
	}
	maxThreads, err := cmd.Flags().GetInt(flagMaxThreadsPerAgent)
	if err != nil {
		return portal.AgentLimits{}, fmt.Errorf("missing max-threads-per-agent flag: %w", err)
	}
	defaultImage, err := cmd.Flags().GetString(flagAgentDefaultImage)
	if err != nil {
		return portal.AgentLimits{}, fmt.Errorf("missing agent-default-image flag: %w", err)
	}
	defaultStorageClass, err := cmd.Flags().GetString(flagAgentDefaultStorageClass)
	if err != nil {
		return portal.AgentLimits{}, fmt.Errorf("missing agent-default-storage-class flag: %w", err)
	}

	return portal.AgentLimits{
		MaxCPU:              maxCPU,
		MaxMemory:           maxMemory,
		MaxThreads:          maxThreads,
		DefaultImage:        defaultImage,
		DefaultStorageClass: defaultStorageClass,
	}, nil
}
