package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/controller"
	awsprovider "github.com/chanzuckerberg/aws-oidc/internal/providers/aws"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
)

const (
	flagOktaAppClientID    = "okta-app-client-id"
	flagIssuerHost         = "issuer-host"
	flagBoundaryPolicyName = "boundary-policy-name"
	flagLeaderElection     = "leader-election"
	flagHealthProbeAddr    = "health-probe-bind-address"
)

func init() {
	rootCmd.AddCommand(operatorCmd)
	operatorCmd.Flags().String(flagOktaAppClientID, "", "Client id of the shared agent Okta app used as the IAM trust audience (TODO: app not created yet)")
	operatorCmd.Flags().String(flagIssuerHost, "czi.okta.com", "Okta issuer host (without scheme) that names the account OIDC provider")
	operatorCmd.Flags().String(flagBoundaryPolicyName, "", "Permissions boundary policy name applied to every agent role (TODO: boundary not created yet; empty skips it)")
	operatorCmd.Flags().Bool(flagLeaderElection, true, "Enable leader election so only one operator replica reconciles")
	operatorCmd.Flags().String(flagHealthProbeAddr, ":8081", "Address the health probe endpoint binds to")
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

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

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

	// AWS is the only provider today. Additional providers are appended here as they are
	// implemented; the reconciler dispatches each grant to the one that handles it.
	awsProvider := awsprovider.NewProvider(awsprovider.Config{
		IssuerHost:         issuerHost,
		OktaAppClientID:    oktaAppClientID,
		BoundaryPolicyName: boundaryPolicyName,
	}, nil)

	reconciler := &controller.AgentReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Providers: []controller.Provider{awsProvider},
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
