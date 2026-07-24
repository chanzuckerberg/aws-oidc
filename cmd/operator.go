package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(operatorCmd)
}

// operatorCmd runs the Agent controller-manager.
//
// STUB: the manager is not wired yet. When implemented this starts a controller-runtime
// manager with leader election that watches Agent CRs and reconciles them through
// internal/controller and internal/provisioner. It reuses the same image as serve-config
// and serve-portal, selected by this subcommand.
var operatorCmd = &cobra.Command{
	Use:           "operator",
	Short:         "aws-oidc operator",
	Long:          "Run the agent-registry operator: watch Agent custom resources and reconcile them into per-agent IAM roles.",
	SilenceErrors: true,
	RunE:          operatorRun,
}

func operatorRun(cmd *cobra.Command, args []string) error {
	// TODO: build a controller-runtime manager (scheme with api/v1 registered, leader
	// election, health/ready probes, a rate-limited workqueue), construct the reconciler
	// from internal/controller with a real internal/provisioner, register it with
	// SetupWithManager, and block on mgr.Start(ctx).
	return fmt.Errorf("operator: not implemented")
}
