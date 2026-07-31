package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
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
		IssuerHost:         issuerHost,
		OktaAppClientID:    oktaAppClientID,
		BoundaryPolicyName: boundaryPolicyName,
		DefaultTags:        roleTags,
	}, clientFactory)

	reconciler := &controller.AgentReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		Providers:           []controller.Provider{awsProvider},
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
