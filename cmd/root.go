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
		return env, errors.Wrap(err, "Unable to load all the environment variables")
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

		err = configureLogrusHooks()
		if err != nil {
			return errors.Wrap(err, "Unable to configure Logrus Hooks")
		}

		err = configureHoneycombTelemetry()
		if err != nil {
			return errors.Wrap(err, "Unable to set up Honeycomb Telemetry")
		}

		return nil
	},
}

func configureLogrusHooks() error {
	sentryEnv, err := loadSentryEnv()
	if err != nil {
		return err
	}
	// if env var not set, ignore
	if sentryEnv.DSN == "" {
		logrus.Debug("Sentry DSN not set. Skipping Sentry Configuration")
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

func configureHoneycombTelemetry() error {
	honeycombEnv, err := loadHoneycombEnv()
	if err != nil {
		return err
	}
	// if env var not set, ignore
	if honeycombEnv.SECRET_KEY == "" {
		logrus.Debug("Honeycomb Secret Key not set. Skipping Honeycomb Configuration")
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
