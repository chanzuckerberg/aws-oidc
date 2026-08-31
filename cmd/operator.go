package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/agentdefaults"
	"github.com/chanzuckerberg/aws-oidc/internal/agentpod"
	"github.com/chanzuckerberg/aws-oidc/internal/controller"
	awsprovider "github.com/chanzuckerberg/aws-oidc/internal/providers/aws"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
)

const (
	flagOktaAppClientID     = "okta-app-client-id"
	flagIssuerHost          = "issuer-host"
	flagBoundaryPolicyName  = "boundary-policy-name"
	flagLeaderElection      = "leader-election"
	flagHealthProbeAddr     = "health-probe-bind-address"
	flagProvisionerRoleName = "provisioner-role-name"
	flagSessionDuration     = "assume-role-session-duration"
	flagAWSRegion           = "aws-region"
	flagGrantConcurrency    = "grant-concurrency"
	flagRoleTags            = "role-tags"
	flagClusterOIDCProvider = "cluster-oidc-provider"
	flagAgentImage          = "agent-default-image"
	flagAgentCommand        = "agent-default-command"
	flagAgentStorageClass   = "agent-storage-class"
	flagMaxThreadsPerAgent  = "max-threads-per-agent"

	flagAnthropicFederationRuleID = "anthropic-federation-rule-id"
	flagAnthropicOrganizationID   = "anthropic-organization-id"
	flagAnthropicServiceAccountID = "anthropic-service-account-id"
	flagAnthropicTokenAudience    = "anthropic-token-audience"

	flagGitHubAppID                  = "github-app-id"
	flagGitHubAppInstallationID      = "github-app-installation-id"
	flagGitHubAppPrivateKeySecret    = "github-app-private-key-secret"
	flagGitHubAppPrivateKeySecretKey = "github-app-private-key-secret-key"
	flagGitHubAPIURL                 = "github-api-url"

	flagDefaultsConfig = "defaults-config"
)

func init() {
	rootCmd.AddCommand(operatorCmd)
	operatorCmd.Flags().String(flagOktaAppClientID, "", "Client id of the shared agent Okta app used as the IAM trust audience")
	operatorCmd.Flags().String(flagIssuerHost, "czi.okta.com", "Okta issuer host (without scheme) that names the account OIDC provider")
	operatorCmd.Flags().String(flagBoundaryPolicyName, "", "Permissions boundary policy name applied to every agent role (TODO: boundary not created yet; empty skips it)")
	operatorCmd.Flags().Bool(flagLeaderElection, true, "Enable leader election so only one operator replica reconciles")
	operatorCmd.Flags().String(flagHealthProbeAddr, ":8081", "Address the health probe endpoint binds to")
	operatorCmd.Flags().String(flagProvisionerRoleName, "agent-provisioner", "Name of the role the operator assumes in each target account to create agent roles")
	operatorCmd.Flags().Duration(flagSessionDuration, 15*time.Minute, "Duration of the short-lived cross-account assume-role session")
	operatorCmd.Flags().String(flagAWSRegion, "us-east-1", "Region used for STS and IAM calls")
	operatorCmd.Flags().Int(flagGrantConcurrency, 8, "Maximum grants of one agent provisioned in parallel")
	operatorCmd.Flags().StringToString(flagRoleTags, nil, "Standard tags applied to every agent role (for example project=agent-registry,env=rdev,service=aws-oidc)")
	operatorCmd.Flags().String(flagClusterOIDCProvider, "", "EKS cluster OIDC issuer without scheme (for example oidc.eks.us-west-2.amazonaws.com/id/EXAMPLE); empty means agents do not run in the cluster")
	operatorCmd.Flags().String(flagAgentImage, "", "Container image an agent thread runs when the agent does not name one")
	operatorCmd.Flags().StringSlice(flagAgentCommand, []string{"sleep", "infinity"}, "Command an agent thread runs when neither the agent nor its image provides a long-running entrypoint")
	operatorCmd.Flags().String(flagAgentStorageClass, "efs-agent-workspaces", "Storage class for the per-agent workspace PVC; must be a ReadWriteMany class backed by the EFS CSI driver")
	operatorCmd.Flags().Int(flagMaxThreadsPerAgent, 5, "Maximum threads one agent may run")
	operatorCmd.Flags().String(flagAnthropicFederationRuleID, os.Getenv("ANTHROPIC_FEDERATION_RULE_ID"), "Anthropic WIF federation rule ID (fdrl_...); when set with the other three anthropic flags, every thread pod receives a Claude WIF token and the four ANTHROPIC_* env vars")
	operatorCmd.Flags().String(flagAnthropicOrganizationID, os.Getenv("ANTHROPIC_ORGANIZATION_ID"), "Anthropic organization UUID (from Settings > Organization in the Claude Console)")
	operatorCmd.Flags().String(flagAnthropicServiceAccountID, os.Getenv("ANTHROPIC_SERVICE_ACCOUNT_ID"), "Anthropic service account ID (svac_...) the minted Claude token acts as")
	operatorCmd.Flags().String(flagAnthropicTokenAudience, os.Getenv("ANTHROPIC_TOKEN_AUDIENCE"), "Audience claim the projected Anthropic token carries; must match the federation rule's audience matcher")
	operatorCmd.Flags().String(flagGitHubAppID, os.Getenv("GITHUB_APP_ID"), "Numeric app ID of the shared GitHub App agents clone and open pull requests as; when set with the installation id and key secret, every thread pod receives the app's credentials")
	operatorCmd.Flags().String(flagGitHubAppInstallationID, os.Getenv("GITHUB_APP_INSTALLATION_ID"), "Installation ID of that GitHub App on the organization whose repositories agents may reach")
	operatorCmd.Flags().String(flagGitHubAppPrivateKeySecret, os.Getenv("GITHUB_APP_PRIVATE_KEY_SECRET"), "Name of the Kubernetes Secret in the operator's namespace holding the GitHub App's PEM private key")
	operatorCmd.Flags().String(flagGitHubAppPrivateKeySecretKey, "private-key.pem", "Key inside that Secret holding the PEM private key")
	operatorCmd.Flags().String(flagGitHubAPIURL, "", "GitHub API base URL; empty means https://api.github.com")
	operatorCmd.Flags().String(flagDefaultsConfig, os.Getenv("AGENT_DEFAULTS_CONFIG"), "Path to the agent-defaults YAML file mounted from the agent-defaults ConfigMap; when set, overrides static default flags without a restart")
}

// operatorCmd runs the Agent controller-manager: it watches Agent custom resources and
// reconciles their grants into provisioned access through the registered providers (AWS
// today). It reuses the same image as serve-config and serve-portal, selected by this
// subcommand.
var operatorCmd = &cobra.Command{
	Use:           "operator",
	Short:         "aws-oidc operator",
	Long:          "Run the agent-registry operator: watch Agent custom resources and reconcile them into provisioned access.",
	SilenceErrors: true,
	RunE:          operatorRun,
}

func operatorRun(cmd *cobra.Command, args []string) error {
	oktaAppClientID, err := cmd.Flags().GetString(flagOktaAppClientID)
	if err != nil {
		return fmt.Errorf("missing okta-app-client-id flag: %w", err)
	}
	issuerHost, err := cmd.Flags().GetString(flagIssuerHost)
	if err != nil {
		return fmt.Errorf("missing issuer-host flag: %w", err)
	}
	boundaryPolicyName, err := cmd.Flags().GetString(flagBoundaryPolicyName)
	if err != nil {
		return fmt.Errorf("missing boundary-policy-name flag: %w", err)
	}
	leaderElection, err := cmd.Flags().GetBool(flagLeaderElection)
	if err != nil {
		return fmt.Errorf("missing leader-election flag: %w", err)
	}
	healthAddr, err := cmd.Flags().GetString(flagHealthProbeAddr)
	if err != nil {
		return fmt.Errorf("missing health-probe-bind-address flag: %w", err)
	}
	provisionerRoleName, err := cmd.Flags().GetString(flagProvisionerRoleName)
	if err != nil {
		return fmt.Errorf("missing provisioner-role-name flag: %w", err)
	}
	sessionDuration, err := cmd.Flags().GetDuration(flagSessionDuration)
	if err != nil {
		return fmt.Errorf("missing assume-role-session-duration flag: %w", err)
	}
	awsRegion, err := cmd.Flags().GetString(flagAWSRegion)
	if err != nil {
		return fmt.Errorf("missing aws-region flag: %w", err)
	}
	grantConcurrency, err := cmd.Flags().GetInt(flagGrantConcurrency)
	if err != nil {
		return fmt.Errorf("missing grant-concurrency flag: %w", err)
	}
	roleTags, err := cmd.Flags().GetStringToString(flagRoleTags)
	if err != nil {
		return fmt.Errorf("missing role-tags flag: %w", err)
	}
	clusterOIDCProvider, err := cmd.Flags().GetString(flagClusterOIDCProvider)
	if err != nil {
		return fmt.Errorf("missing cluster-oidc-provider flag: %w", err)
	}
	agentImage, err := cmd.Flags().GetString(flagAgentImage)
	if err != nil {
		return fmt.Errorf("missing agent-default-image flag: %w", err)
	}
	agentCommand, err := cmd.Flags().GetStringSlice(flagAgentCommand)
	if err != nil {
		return fmt.Errorf("missing agent-default-command flag: %w", err)
	}
	agentStorageClass, err := cmd.Flags().GetString(flagAgentStorageClass)
	if err != nil {
		return fmt.Errorf("missing agent-storage-class flag: %w", err)
	}
	maxThreads, err := cmd.Flags().GetInt(flagMaxThreadsPerAgent)
	if err != nil {
		return fmt.Errorf("missing max-threads-per-agent flag: %w", err)
	}
	anthropicFederationRuleID, err := cmd.Flags().GetString(flagAnthropicFederationRuleID)
	if err != nil {
		return fmt.Errorf("missing anthropic-federation-rule-id flag: %w", err)
	}
	anthropicOrganizationID, err := cmd.Flags().GetString(flagAnthropicOrganizationID)
	if err != nil {
		return fmt.Errorf("missing anthropic-organization-id flag: %w", err)
	}
	anthropicServiceAccountID, err := cmd.Flags().GetString(flagAnthropicServiceAccountID)
	if err != nil {
		return fmt.Errorf("missing anthropic-service-account-id flag: %w", err)
	}
	anthropicTokenAudience, err := cmd.Flags().GetString(flagAnthropicTokenAudience)
	if err != nil {
		return fmt.Errorf("missing anthropic-token-audience flag: %w", err)
	}
	githubAppID, err := cmd.Flags().GetString(flagGitHubAppID)
	if err != nil {
		return fmt.Errorf("missing github-app-id flag: %w", err)
	}
	githubAppInstallationID, err := cmd.Flags().GetString(flagGitHubAppInstallationID)
	if err != nil {
		return fmt.Errorf("missing github-app-installation-id flag: %w", err)
	}
	githubAppPrivateKeySecret, err := cmd.Flags().GetString(flagGitHubAppPrivateKeySecret)
	if err != nil {
		return fmt.Errorf("missing github-app-private-key-secret flag: %w", err)
	}
	githubAppPrivateKeySecretKey, err := cmd.Flags().GetString(flagGitHubAppPrivateKeySecretKey)
	if err != nil {
		return fmt.Errorf("missing github-app-private-key-secret-key flag: %w", err)
	}
	githubAPIURL, err := cmd.Flags().GetString(flagGitHubAPIURL)
	if err != nil {
		return fmt.Errorf("missing github-api-url flag: %w", err)
	}
	defaultsConfigPath, err := cmd.Flags().GetString(flagDefaultsConfig)
	if err != nil {
		return fmt.Errorf("missing defaults-config flag: %w", err)
	}

	// Route controller-runtime's logr logging through the repo's slog logger set up in
	// PersistentPreRunE, so the operator logs the same way as the rest of the binary.
	ctrl.SetLogger(logr.FromSlogHandler(slog.Default().Handler()))

	restConfig, namespace, err := configmap.NewInClusterConfig()
	if err != nil {
		return fmt.Errorf("loading in-cluster config: %w", err)
	}

	scheme := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("registering client-go scheme: %w", err)
	}
	err = agentsv1.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("registering agents scheme: %w", err)
	}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                  scheme,
		LeaderElection:          leaderElection,
		LeaderElectionID:        "aws-oidc-agent-operator",
		LeaderElectionNamespace: namespace,
		HealthProbeBindAddress:  healthAddr,
		Metrics:                 metricsserver.Options{BindAddress: "0"},
		Cache:                   cache.Options{ByObject: threadObjectCache(namespace)},
	})
	if err != nil {
		return fmt.Errorf("creating manager: %w", err)
	}

	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	if err != nil {
		return fmt.Errorf("adding health check: %w", err)
	}
	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	if err != nil {
		return fmt.Errorf("adding ready check: %w", err)
	}

	// The operator's own (IRSA) identity assumes each target account's provisioner role to
	// create agent roles there, with short-lived sessions.
	clientFactory, err := awsprovider.NewAssumeRoleClientFactory(cmd.Context(), awsprovider.AssumeRoleConfig{
		ProvisionerRoleName: provisionerRoleName,
		SessionDuration:     sessionDuration,
		Region:              awsRegion,
	})
	if err != nil {
		return fmt.Errorf("building cross-account client factory: %w", err)
	}

	// AWS is the only provider today. Additional providers are appended here as they are
	// implemented; the reconciler dispatches each grant to the one that handles it.
	awsProvider := awsprovider.NewProvider(awsprovider.Config{
		IssuerHost:          issuerHost,
		OktaAppClientID:     oktaAppClientID,
		BoundaryPolicyName:  boundaryPolicyName,
		DefaultTags:         roleTags,
		ClusterOIDCProvider: clusterOIDCProvider,
		PodNamespace:        namespace,
	}, clientFactory)

	// Threads only run where the cluster's OIDC issuer is configured, since without it their
	// pods have no way to assume the agent's roles.
	var threads controller.ThreadReconciler
	if clusterOIDCProvider != "" {
		threads = agentpod.New(mgr.GetClient(), mgr.GetScheme(), agentpod.Config{
			Namespace:                 namespace,
			DefaultsLoader:            agentdefaults.NewLoader(defaultsConfigPath),
			DefaultImage:              agentImage,
			DefaultCommand:            agentCommand,
			StorageClass:              agentStorageClass,
			Region:                    awsRegion,
			MaxThreads:                maxThreads,
			AnthropicFederationRuleID: anthropicFederationRuleID,
			AnthropicOrganizationID:   anthropicOrganizationID,
			AnthropicServiceAccountID: anthropicServiceAccountID,
			AnthropicTokenAudience:    anthropicTokenAudience,

			GitHubAppID:                  githubAppID,
			GitHubAppInstallationID:      githubAppInstallationID,
			GitHubAppPrivateKeySecret:    githubAppPrivateKeySecret,
			GitHubAppPrivateKeySecretKey: githubAppPrivateKeySecretKey,
			GitHubAPIURL:                 githubAPIURL,
		})
	}

	reconciler := &controller.AgentReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		Providers:           []controller.Provider{awsProvider},
		Threads:             threads,
		MaxConcurrentGrants: grantConcurrency,
	}
	err = reconciler.SetupWithManager(mgr)
	if err != nil {
		return fmt.Errorf("setting up reconciler: %w", err)
	}

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		return fmt.Errorf("running manager: %w", err)
	}
	return nil
}

// threadObjectCache confines the caches for the objects an agent's threads are made of to the
// operator's own namespace. Agents are still watched cluster-wide, but the operator's access to
// workloads is deliberately namespaced, so a cluster-wide informer for these would be refused
// and the controller would never start.
func threadObjectCache(namespace string) map[client.Object]cache.ByObject {
	own := cache.ByObject{Namespaces: map[string]cache.Config{namespace: {}}}
	return map[client.Object]cache.ByObject{
		&appsv1.StatefulSet{}:           own,
		&corev1.ServiceAccount{}:        own,
		&corev1.Service{}:               own,
		&corev1.ConfigMap{}:             own,
		&corev1.PersistentVolumeClaim{}: own,
	}
}
