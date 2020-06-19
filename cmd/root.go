package cmd

import (
	"context"

	"github.com/evalphobia/logrus_sentry"
	"github.com/honeycombio/beeline-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clientID string
var issuerURL string
var roleARN string

const (
	flagVerbose = "verbose"
)

type SentryEnvironment struct {
	DSN string
}

func loadSentryEnv() (*SentryEnvironment, error) {
	env := &SentryEnvironment{}
	err := envconfig.Process("SENTRY", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the environment variables")
	}
	return env, nil
}

type HoneycombEnvironment struct {
	SECRET_KEY string
}

func loadHoneycombEnv() (*HoneycombEnvironment, error) {
	env := &HoneycombEnvironment{}
	err := envconfig.Process("HONEYCOMB", env)
	if err != nil {
		return env, errors.Wrap(err, "Unable to load all the honeycomb environment variables")
	}
	return env, nil
}

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
			log.SetReportCaller(true)
		}
		// Configure Logrus
		err = configureLogrusHooks()
		if err != nil {
			return errors.Wrap(err, "Unable to configure Logrus Hooks")
		}
		// Configure Honeycomb Telemetry
		honeycombEnv, err := loadHoneycombEnv()
		if err != nil {
			return err
		}
		beeline.Init(beeline.Config{
			WriteKey:    honeycombEnv.SECRET_KEY,
			Dataset:     "aws-oidc",
			ServiceName: "aws-oidc",
		})
		return nil
	},
}

func configureLogrusHooks() error {
	// Load Sentry Env
	sentryEnv, err := loadSentryEnv()
	if err != nil {
		return err
	}
	// if env var not set, ignore
	if sentryEnv.DSN == "" {
		return nil
	}

	sentryHook, err := logrus_sentry.NewSentryHook(sentryEnv.DSN, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})
	if err != nil {
		logrus.Errorf("Error configuring Sentry")
		return nil
	}
	log.AddHook(sentryHook)
	return nil
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
