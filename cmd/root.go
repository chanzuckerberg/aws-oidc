package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/honeycombio/beeline-go"
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

type HoneycombEnvironment struct {
	SECRET_KEY   string
	DATASET_NAME string `default:"aws-oidc"`
	SERVICE_NAME string `default:"aws-oidc"`
}

func loadHoneycombEnv() (*HoneycombEnvironment, error) {
	env := &HoneycombEnvironment{}
	err := envconfig.Process("HONEYCOMB", env)
	if err != nil {
		return env, fmt.Errorf("Unable to load all the honeycomb environment variables: %w", err)
	}
	return env, nil
}

func init() {
	rootCmd.PersistentFlags().BoolP(flagVerbose, "v", false, "Use this to enable verbose mode")
}

func initLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

var rootCmd = &cobra.Command{
	Use:          "aws-oidc",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initLogger()
		// parse flags
		verbose, err := cmd.Flags().GetBool(flagVerbose)
		if err != nil {
			return fmt.Errorf("Missing verbose flag: %w", err)
		}
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		err = configureLogHooks()
		if err != nil {
			return fmt.Errorf("Unable to configure log Hooks: %w", err)
		}

		err = configureHoneycombTelemetry()
		if err != nil {
			return fmt.Errorf("Unable to set up Honeycomb Telemetry: %w", err)
		}

		return nil
	},
}

func configureLogHooks() error {
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
}

func configureHoneycombTelemetry() error {
	honeycombEnv, err := loadHoneycombEnv()
	if err != nil {
		return err
	}
	// if env var not set, ignore
	if honeycombEnv.SECRET_KEY == "" {
		slog.Debug("Honeycomb Secret Key not set. Skipping Honeycomb Configuration")
		return nil
	}
	beeline.Init(beeline.Config{
		WriteKey:    honeycombEnv.SECRET_KEY,
		Dataset:     honeycombEnv.DATASET_NAME,
		ServiceName: honeycombEnv.SERVICE_NAME,
	})

	return nil
}

func Execute(ctx context.Context) error {
	defer beeline.Close()
	return rootCmd.ExecuteContext(ctx)
}
