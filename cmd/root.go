package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var clientID string
var issuerURL string
var roleARN string

const (
	flagVerbose             = "verbose"
	flagFlushOIDCTokenCache = "flush-oidc-token-cache"
	flagLogFile             = "log-file"
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
var logCloser func() error = func() error { return nil }

func getDefaultLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/aws-oidc.log"
	}
	return filepath.Join(homeDir, ".aws-oidc", "logs", "aws-oidc.log")
}

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
	rootCmd.PersistentFlags().BoolP(flagFlushOIDCTokenCache, "", false, "Flush the OIDC token cache")
	rootCmd.PersistentFlags().BoolVar(&deviceCodeFlow, "device-code-flow", false, "Use device code flow for authentication")
	rootCmd.PersistentFlags().String(flagLogFile, "", "Path to a log file to write logs to (in addition to stderr when verbose)")
}

func initLogger(verbose bool, logFile string) (func() error, error) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	var writers []io.Writer = []io.Writer{
		os.Stderr,
	}

	if logFile == "" {
		logFile = getDefaultLogFile()
	}

	logDir := filepath.Dir(logFile)
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	writers = append(writers, f)
	closer := func() error {
		return f.Close()
	}

	writer := io.MultiWriter(writers...)
	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return closer, nil
}

func flushOIDCTokenCacheFn(ctx context.Context, clientID, issuerURL string) error {
	storage, err := storage.GetOIDC(clientID, issuerURL)
	if err != nil {
		return fmt.Errorf("getting oidc token storage: %w", err)
	}

	err = storage.Delete(ctx)
	if err != nil {
		return fmt.Errorf("deleting token from storage: %w", err)
	}

	return nil
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
		logFile, err := cmd.Flags().GetString(flagLogFile)
		if err != nil {
			return fmt.Errorf("missing log-file flag: %w", err)
		}
		if flushOIDCTokenCache {
			err = flushOIDCTokenCacheFn(cmd.Context(), clientID, issuerURL)
			if err != nil {
				return fmt.Errorf("flushing oidc token cache: %w", err)
			}
		}

		logCloser, err = initLogger(verbose, logFile)
		if err != nil {
			return fmt.Errorf("initializing logger: %w", err)
		}

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
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s failed: %v", cmd.CommandPath(), err))
	}
	closeErr := logCloser()
	if closeErr != nil {
		fmt.Fprintf(os.Stderr, "closing log file: %v\n", closeErr)
	}
	return err
}
