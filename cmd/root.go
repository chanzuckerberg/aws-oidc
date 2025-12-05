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
	flagVerbose             = "verbose"
	flagFlushOIDCTokenCache = "flush-oidc-token-cache"
	successMessage          = `<h1>Success!</h1><p>You are now authenticated with AWS; this temporary session
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
		return env, fmt.Errorf("loading all the environment variables: %w", err)
	}
	return env, nil
}

var deviceCodeFlow bool

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
	rootCmd.PersistentFlags().BoolP(flagFlushOIDCTokenCache, "", false, "Flush the OIDC token cache")
	rootCmd.Flags().BoolVar(&deviceCodeFlow, "device-code-flow", false, "Use device code flow for authentication")
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
		verbose, err := cmd.Flags().GetBool(flagVerbose)
		if err != nil {
			return fmt.Errorf("missing verbose flag: %w", err)
		}
		flushOIDCTokenCache, err := cmd.Flags().GetBool(flagFlushOIDCTokenCache)
		if err != nil {
			return fmt.Errorf("missing flush-oidc-token-cache flag: %w", err)
		}
		if flushOIDCTokenCache {
			err = flushOIDCTokenCacheFn(cmd.Context(), clientID, issuerURL)
			if err != nil {
				return fmt.Errorf("flushing oidc token cache: %w", err)
			}
			return nil
		}

		initLogger(verbose)

		sentryEnv, err := loadSentryEnv()
		if err != nil {
			return err
		}
		// if env var not set, ignore
		if sentryEnv.DSN == "" {
			slog.Debug("Sentry DSN not set. Skipping Sentry Configuration")
		}

		return nil
	},
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
