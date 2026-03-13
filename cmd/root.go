package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/storage"
	"github.com/spf13/cobra"
)

var (
	clientID               string
	issuerURL              string
	roleARN                string
	deviceCodeFlow         bool
	nodeLocalCache         string
	nodeLocalCacheFullPath string
	logCloser              func() error = func() error { return nil }
)

const (
	flagVerbose             = "verbose"
	flagFlushOIDCTokenCache = "flush-oidc-token-cache"
	flagLogLokiURL          = "log-loki-url"
	flagLogLokiCredentials  = "log-loki-credentials"
	flagNodeLocalCache      = "node-local-cache"
	envNodeLocalCache       = "AWS_OIDC_NODE_LOCAL_CACHE"
	envLogLokiURL           = "AWS_OIDC_LOG_LOKI_URL"
	envLogLokiCredentials   = "AWS_OIDC_LOG_LOKI_CREDENTIALS" // base64-encoded username:password
	successMessage          = `<h1>Success!</h1><p>You are now authenticated with AWS; this temporary session
	will allow you to run AWS commmands from the command line.</p><p> When running
	aws-cli commands, be sure to specify your profile in one of the following ways:</p>
	<code>$ aws --profile &lt;profile-name&gt; &lt;command&gt;</code><br/>
	<code>$ AWS_PROFILE=&lt;profile-name&gt; aws &lt;command&gt;</code><br/>
	<p> Feel free to <a href="#" onclick="window.close();">close this window</a> whenever</p>
	`
	defaultLokiURL         = "https://public-loki.prod-central-o11y.prod.czi.team"
	defaultLokiCredentials = ""
)

func init() {
	rootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity (-v for INFO, -vv for DEBUG)")
	rootCmd.PersistentFlags().BoolP(flagFlushOIDCTokenCache, "", false, "Flush the OIDC token cache")
	rootCmd.PersistentFlags().BoolVar(&deviceCodeFlow, "device-code-flow", false, "Use device code flow for authentication")

	logLokiURL := os.Getenv(envLogLokiURL)
	if logLokiURL == "" {
		logLokiURL = defaultLokiURL
	}
	rootCmd.PersistentFlags().String(flagLogLokiURL, logLokiURL, "Loki base URL to push logs to")

	logLokiCredentials := os.Getenv(envLogLokiCredentials)
	if logLokiCredentials == "" {
		logLokiCredentials = defaultLokiCredentials
	}
	rootCmd.PersistentFlags().String(flagLogLokiCredentials, logLokiCredentials, "Loki credentials to push logs to")

	rootCmd.PersistentFlags().StringVar(&nodeLocalCache, flagNodeLocalCache, os.Getenv(envNodeLocalCache),
		"Directory on node-local disk for OIDC token cache (use in distributed/NFS environments). "+
			"Can also be set via "+envNodeLocalCache)
}

func flushOIDCTokenCacheFn(ctx context.Context, clientID, issuerURL string) error {
	storage, err := storage.GetOIDC(ctx, clientID, issuerURL)
	if err != nil {
		return fmt.Errorf("getting oidc token storage: %w", err)
	}

	err = storage.Delete(ctx)
	if err != nil {
		return fmt.Errorf("deleting token from storage: %w", err)
	}

	return nil
}

func getNodeLocalCachePath(provided string) string {
	// hostname is an empty string if the hostname is not found.
	hostname, _ := os.Hostname()
	// Each user has their own directory on the node-local disk.
	return filepath.Join(provided, fmt.Sprintf("%d", os.Getuid()), hostname)
}

var rootCmd = &cobra.Command{
	Use:          "aws-oidc",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbosity, err := cmd.Flags().GetCount(flagVerbose)
		if err != nil {
			return fmt.Errorf("missing verbose flag: %w", err)
		}

		flushOIDCTokenCache, err := cmd.Flags().GetBool(flagFlushOIDCTokenCache)
		if err != nil {
			return fmt.Errorf("missing flush-oidc-token-cache flag: %w", err)
		}

		logLokiURL, err := cmd.Flags().GetString(flagLogLokiURL)
		if err != nil {
			return fmt.Errorf("missing log-loki-url flag: %w", err)
		}

		logLokiCredentials, err := cmd.Flags().GetString(flagLogLokiCredentials)
		if err != nil {
			return fmt.Errorf("missing log-loki-credentials flag: %w", err)
		}

		if flushOIDCTokenCache {
			err = flushOIDCTokenCacheFn(cmd.Context(), clientID, issuerURL)
			if err != nil {
				return fmt.Errorf("flushing oidc token cache: %w", err)
			}
		}

		if nodeLocalCache != "" {
			nodeLocalCacheFullPath = getNodeLocalCachePath(nodeLocalCache)
		}

		logCloser, err = initLogger(verbosity, logLokiURL, logLokiCredentials)
		if err != nil {
			return fmt.Errorf("initializing logger: %w", err)
		}

		return nil
	},
}

func Execute(ctx context.Context) error {
	cmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %v\n", cmd.CommandPath(), err)
	}
	closeErr := logCloser()
	if closeErr != nil {
		fmt.Fprintf(os.Stderr, "flushing logger: %v\n", closeErr)
	}
	return err
}
