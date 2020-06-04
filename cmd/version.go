package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:           "version",
	Short:         "aws-oidc version",
	Long:          "current aws-oidc version installed",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := util.VersionString()
		if err != nil {
			return err
		}
		fmt.Println(version)
		return nil
	},
}
