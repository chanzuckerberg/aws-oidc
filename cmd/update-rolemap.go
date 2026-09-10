package cmd

import (
	"fmt"
	"log/slog"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chanzuckerberg/aws-oidc/pkg/configmap"
	"github.com/chanzuckerberg/aws-oidc/pkg/rolemap"
	"github.com/spf13/cobra"
)

const (
	flagConfigMapName = "configmap-name"
	flagConfigMapKey  = "configmap-key"
	_flagStateBucket  = "state-bucket"
	_flagStatePrefix  = "state-prefix"
	_flagStateRegion  = "state-region"

	_defaultStateBucket = "terragrunt-engine-state"
	_defaultStatePrefix = "terraform/shared-infra/accounts/"
	_defaultStateRegion = "us-west-2"
)

func init() {
	rootCmd.AddCommand(_updateRolemapCmd)
	_updateRolemapCmd.Flags().String(flagConfigMapName, "rolemap", "Name of the ConfigMap to write the rolemap to")
	_updateRolemapCmd.Flags().String(flagConfigMapKey, "rolemap.yaml", "Key within the ConfigMap to store the rolemap YAML under")
	_updateRolemapCmd.Flags().String(_flagStateBucket, _defaultStateBucket, "S3 bucket containing Terraform state")
	_updateRolemapCmd.Flags().String(_flagStatePrefix, _defaultStatePrefix, "S3 prefix containing account Terraform state")
	_updateRolemapCmd.Flags().String(_flagStateRegion, _defaultStateRegion, "AWS region containing the state bucket")
}

var _updateRolemapCmd = &cobra.Command{
	Use:           "update-rolemap",
	Short:         "aws-oidc update-rolemap",
	Long:          "Generate the rolemap from Terraform state in S3 and write it to the rolemap ConfigMap. Runs in-cluster on a schedule.",
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
	stateBucket, err := cmd.Flags().GetString(_flagStateBucket)
	if err != nil {
		return fmt.Errorf("reading state-bucket flag: %w", err)
	}
	statePrefix, err := cmd.Flags().GetString(_flagStatePrefix)
	if err != nil {
		return fmt.Errorf("reading state-prefix flag: %w", err)
	}
	stateRegion, err := cmd.Flags().GetString(_flagStateRegion)
	if err != nil {
		return fmt.Errorf("reading state-region flag: %w", err)
	}

	ctx := cmd.Context()
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(stateRegion))
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}

	mappings, err := rolemap.Generate(ctx, s3.NewFromConfig(awsConfig), stateBucket, statePrefix)
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
