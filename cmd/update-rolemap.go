package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/chanzuckerberg/aws-oidc/pkg/rolemap"
	"github.com/spf13/cobra"
)

const (
	flagConfigMapName = "configmap-name"
	flagConfigMapKey  = "configmap-key"
)

func init() {
	rootCmd.AddCommand(updateRolemapCmd)
	updateRolemapCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to write the rolemap to")
	updateRolemapCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap to store the rolemap YAML under")
}

var updateRolemapCmd = &cobra.Command{
	Use:           "update-rolemap",
	Short:         "aws-oidc update-rolemap",
	Long:          "Generate the rolemap from TFE state and write it to the rolemap ConfigMap. Runs in-cluster on a schedule.",
	SilenceErrors: true,
	RunE:          updateRolemapRun,
}

func updateRolemapRun(cmd *cobra.Command, args []string) error {
	configMapName, err := cmd.Flags().GetString(flagConfigMapName)
	if err != nil {
		return fmt.Errorf("missing configmap-name flag: %w", err)
	}
	configMapKey, err := cmd.Flags().GetString(flagConfigMapKey)
	if err != nil {
		return fmt.Errorf("missing configmap-key flag: %w", err)
	}

	tfeToken := os.Getenv("TFE_TOKEN")
	if tfeToken == "" {
		return fmt.Errorf("TFE_TOKEN environment variable must be set")
	}

	ctx := cmd.Context()
	mappings, err := rolemap.Generate(ctx, tfeToken)
	if err != nil {
		return fmt.Errorf("generating rolemap: %w", err)
	}

	data, err := rolemap.Marshal(mappings)
	if err != nil {
		return err
	}

	client, namespace, err := configmap.NewInClusterClient()
	if err != nil {
		return fmt.Errorf("creating in-cluster client: %w", err)
	}

	err = configmap.WriteData(ctx, client, namespace, configMapName, configMapKey, data)
	if err != nil {
		return fmt.Errorf("writing rolemap configmap: %w", err)
	}

	slog.Info("wrote rolemap", "namespace", namespace, "configmap", configMapName, "mappings", len(mappings))
	return nil
}
