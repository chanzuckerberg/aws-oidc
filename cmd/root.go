package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var clientID string
var issuerURL string
var roleARN string

const (
	flagVerbose    = "verbose"
	successMessage = `<h1>Success!</h1><p>You are now authenticated with AWS; this temporary session
	will allow you to run AWS commmands from the command line.</p><p> When running
	aws-cli commands, be sure to specify your profile in one of the following ways:</p>
	<code>$ aws --profile &lt;profile-name&gt; &lt;command&gt;</code><br/>
	<code>$ AWS_PROFILE=&lt;profile-name&gt; aws &lt;command&gt;</code><br/>
	<p> Feel free to <a href="#" onclick="window.close();">close this window</a> whenever</p>
	`
)

type SentryEnvironment struct {
	DSN string
}

func loadSentryEnv() (*SentryEnvironment, error) {
	env := &SentryEnvironment{}
	err := envconfig.Process("SENTRY", env)
	if err != nil {
		return env, fmt.Errorf("Unable to load all the environment variables: %w", err)
	}
	return env, nil
}

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
}

func initLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}

var rootCmd = &cobra.Command{
	Use:          "aws-oidc",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// parse flags
		verbose, err := cmd.Flags().GetBool(flagVerbose)
		if err != nil {
			return fmt.Errorf("Missing verbose flag: %w", err)
		}
		initLogger(verbose)

		sentryEnv, err := loadSentryEnv()
		if err != nil {
			return err
		}
		// if env var not set, ignore
		if sentryEnv.DSN == "" {
			slog.Debug("Sentry DSN not set. Skipping Sentry Configuration")
			return nil
		}

		return nil
	},
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
