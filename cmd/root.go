package cmd

import (
	"context"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clientID string
var issuerURL string
var roleARN string

const (
	flagVerbose = "verbose"
)

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
}

var rootCmd = &cobra.Command{
	Use:          "aws-oidc",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// parse flags
		verbose, err := cmd.Flags().GetBool(flagVerbose)
		if err != nil {
			return errors.Wrap(err, "Missing verbose flag")
		}
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	},
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
