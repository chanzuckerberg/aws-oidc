package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/imds_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/workload_oidc"
	"github.com/spf13/cobra"
)

const (
	defaultIMDSPort         = 9911
	defaultBindAddress      = "127.0.0.1"
	defaultRefreshBefore    = 5 * time.Minute
	defaultSessionDuration  = time.Hour
	defaultStartupFetchWait = 30 * time.Second
	shutdownDrainTimeout    = 5 * time.Second
)

var (
	serveIMDSClientID         string
	serveIMDSClientSecretFile string
	serveIMDSIssuerURL        string
	serveIMDSAWSRoleARN       string
	serveIMDSScope            string
	serveIMDSAWSRegion        string
	serveIMDSBindAddress      string
	serveIMDSPort             int
	serveIMDSRefreshBefore    time.Duration
	serveIMDSSessionDuration  time.Duration
	serveIMDSRequireV2        bool
)

func init() {
	serveIMDSCmd.Flags().StringVar(&serveIMDSClientID, "client-id", "",
		"Okta Service app client_id. Or set AWS_OIDC_IMDS_CLIENT_ID.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSClientSecretFile, "client-secret-file", "",
		"Path to a file containing the Okta Service app client_secret. File MUST be mode 0600. Or set AWS_OIDC_IMDS_CLIENT_SECRET_FILE.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSIssuerURL, "issuer-url", "",
		"Okta CUSTOM Authorization Server URL (e.g. https://example.okta.com/oauth2/aus123abc). MUST NOT be the Org/default authorization server. Or set AWS_OIDC_IMDS_ISSUER_URL.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSAWSRoleARN, "aws-role-arn", "",
		"IAM role ARN to assume via sts:AssumeRoleWithWebIdentity. Or set AWS_OIDC_IMDS_AWS_ROLE_ARN.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSScope, "scope", "",
		"OAuth scope to request (typically a custom scope on the Custom AS, e.g. 'aws-m2m-access'). Or set AWS_OIDC_IMDS_SCOPE.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSAWSRegion, "aws-region", "us-west-2",
		"STS regional endpoint region. Or set AWS_OIDC_IMDS_AWS_REGION.")
	serveIMDSCmd.Flags().StringVar(&serveIMDSBindAddress, "bind-address", defaultBindAddress,
		"IP to bind the IMDS listener on. NEVER bind to a non-loopback address unless the helper runs in a sandboxed network namespace (e.g. K8s sidecar in same pod).")
	serveIMDSCmd.Flags().IntVar(&serveIMDSPort, "port", defaultIMDSPort,
		"Port to bind the IMDS listener on. Default 9911 matches AWS's aws_signing_helper.")
	serveIMDSCmd.Flags().DurationVar(&serveIMDSRefreshBefore, "refresh-before", defaultRefreshBefore,
		"Pre-expiry refresh window for STS credentials.")
	serveIMDSCmd.Flags().DurationVar(&serveIMDSSessionDuration, "session-duration", defaultSessionDuration,
		"STS DurationSeconds requested. Capped by the IAM role's MaxSessionDuration; a smaller cap on the role wins.")
	serveIMDSCmd.Flags().BoolVar(&serveIMDSRequireV2, "require-imdsv2", false,
		"Refuse v1-style GET requests that lack the X-Aws-Ec2-Metadata-Token header. Default false (SDK fallback compatibility).")

	rootCmd.AddCommand(serveIMDSCmd)
}

// serveIMDSCmd is the workload-OIDC IMDS-mock subcommand.
var serveIMDSCmd = &cobra.Command{
	Use:   "serve-imds",
	Short: "Run a localhost IMDSv2-shaped credential server backed by Okta workload OIDC",
	Long: `serve-imds runs a long-lived credential helper that:

  1. Authenticates to Okta with a Service app via OAuth 2.0 client_credentials,
     obtaining a JWT access_token from a Custom Authorization Server.
  2. Exchanges the JWT for AWS STS credentials via sts:AssumeRoleWithWebIdentity.
  3. Serves the STS credentials on a localhost endpoint that mimics EC2 IMDSv2.

Workloads such as Prometheus or the AWS CLI can use the served credentials by
setting AWS_EC2_METADATA_SERVICE_ENDPOINT to the helper's URL — the AWS SDK's
default credential chain will discover the endpoint and SigV4-sign requests
with the served credentials transparently.

The --client-secret-file MUST be mode 0600 and contain only the secret
string. Flags can be overridden via env vars with prefix AWS_OIDC_IMDS_.`,
	SilenceErrors: true,
	RunE:          serveIMDSRun,
}

func serveIMDSRun(cmd *cobra.Command, _ []string) error {
	loadServeIMDSEnv()

	cfg, err := buildServeIMDSConfig()
	if err != nil {
		return err
	}

	clientSecret, err := readClientSecretFile(cfg.clientSecretFile)
	if err != nil {
		return err
	}

	oidcCfg := workload_oidc.Config{
		ClientID:     cfg.clientID,
		ClientSecret: clientSecret,
		IssuerURL:    cfg.issuerURL,
		Scopes:       cfg.scopes(),
	}

	roleName, err := roleNameFromARN(cfg.awsRoleARN)
	if err != nil {
		return err
	}

	awsCfg, err := config.LoadDefaultConfig(cmd.Context(),
		config.WithRegion(cfg.awsRegion),
	)
	if err != nil {
		return fmt.Errorf("loading AWS config: %w", err)
	}
	stsClient := sts.NewFromConfig(awsCfg)

	cache, err := imds_server.NewCredentialsCache(imds_server.NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         cfg.awsRoleARN,
		RoleSessionName: cfg.clientID, // sanitized inside getter where used by other paths; cache passes through to STS
		SessionDuration: cfg.sessionDuration,
		Fetcher:         oidcCfg.FetchToken,
		RefreshBefore:   cfg.refreshBefore,
	})
	if err != nil {
		return fmt.Errorf("building credentials cache: %w", err)
	}

	server, err := imds_server.NewServer(imds_server.Options{
		RoleName:      roleName,
		Cache:         cache,
		RequireIMDSv2: cfg.requireIMDSv2,
		Logger:        slog.Default(),
	})
	if err != nil {
		return fmt.Errorf("building imds server: %w", err)
	}

	// Startup validation (fail-fast). Performs one Okta -> STS round-trip
	// before binding the listener so misconfiguration produces a clear,
	// immediate error.
	startupCtx, startupCancel := context.WithTimeout(cmd.Context(), defaultStartupFetchWait)
	defer startupCancel()
	if _, err := cache.Get(startupCtx); err != nil {
		return fmt.Errorf("startup credential fetch failed: %w", err)
	}
	slog.Info("imds_server: startup credential fetch succeeded",
		slog.String("role_arn", cfg.awsRoleARN),
		slog.Any("oidc", oidcCfg),
	)

	// Start the HTTP listener. Warn if binding to non-loopback.
	if !isLoopbackBindAddress(cfg.bindAddress) {
		slog.Warn("imds_server: binding to non-loopback address — anyone with network access to this host can fetch AWS credentials",
			slog.String("bind_address", cfg.bindAddress),
		)
	}
	addr := net.JoinHostPort(cfg.bindAddress, fmt.Sprintf("%d", cfg.port))
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown on SIGTERM / SIGINT.
	rootCtx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	listenErrCh := make(chan error, 1)
	go func() {
		slog.Info("imds_server: listening", slog.String("addr", "http://"+addr))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErrCh <- err
			return
		}
		listenErrCh <- nil
	}()

	select {
	case <-rootCtx.Done():
		slog.Info("imds_server: shutdown signal received, draining")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownDrainTimeout)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			slog.Error("imds_server: shutdown error", slog.String("error", err.Error()))
		}
		<-listenErrCh
		return nil
	case err := <-listenErrCh:
		if err != nil {
			return fmt.Errorf("imds listener: %w", err)
		}
		return nil
	}
}

// serveIMDSConfig is the validated set of inputs for the subcommand.
type serveIMDSConfig struct {
	clientID         string
	clientSecretFile string
	issuerURL        string
	awsRoleARN       string
	scope            string
	awsRegion        string
	bindAddress      string
	port             int
	refreshBefore    time.Duration
	sessionDuration  time.Duration
	requireIMDSv2    bool
}

func (c serveIMDSConfig) scopes() []string {
	if c.scope == "" {
		return nil
	}
	return strings.Fields(c.scope)
}

// LogValue redacts nothing at this level (no secrets in this struct), but
// implements LogValuer for consistent slog formatting.
func (c serveIMDSConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("client_id", c.clientID),
		slog.String("client_secret_file", c.clientSecretFile),
		slog.String("issuer_url", c.issuerURL),
		slog.String("aws_role_arn", c.awsRoleARN),
		slog.String("scope", c.scope),
		slog.String("aws_region", c.awsRegion),
		slog.String("bind_address", c.bindAddress),
		slog.Int("port", c.port),
		slog.Duration("refresh_before", c.refreshBefore),
		slog.Duration("session_duration", c.sessionDuration),
		slog.Bool("require_imdsv2", c.requireIMDSv2),
	)
}

func buildServeIMDSConfig() (serveIMDSConfig, error) {
	if serveIMDSClientID == "" {
		return serveIMDSConfig{}, errors.New("missing required flag/env: --client-id (AWS_OIDC_IMDS_CLIENT_ID)")
	}
	if serveIMDSClientSecretFile == "" {
		return serveIMDSConfig{}, errors.New("missing required flag/env: --client-secret-file (AWS_OIDC_IMDS_CLIENT_SECRET_FILE)")
	}
	if serveIMDSIssuerURL == "" {
		return serveIMDSConfig{}, errors.New("missing required flag/env: --issuer-url (AWS_OIDC_IMDS_ISSUER_URL)")
	}
	if serveIMDSAWSRoleARN == "" {
		return serveIMDSConfig{}, errors.New("missing required flag/env: --aws-role-arn (AWS_OIDC_IMDS_AWS_ROLE_ARN)")
	}
	if !strings.HasPrefix(serveIMDSIssuerURL, "https://") && !strings.HasPrefix(serveIMDSIssuerURL, "http://") {
		return serveIMDSConfig{}, errors.New("--issuer-url must start with http:// or https://")
	}
	// Defensive: catch the common "Org AS" mistake. Org AS is /oauth2/v1
	// rather than /oauth2/<auth_server_id>. The Org AS issues opaque tokens
	// AWS STS cannot validate.
	if strings.HasSuffix(strings.TrimRight(serveIMDSIssuerURL, "/"), "/oauth2/v1") {
		return serveIMDSConfig{}, fmt.Errorf("--issuer-url looks like the Okta Org Authorization Server (%s); a Custom Authorization Server is required (e.g. .../oauth2/aus<id>)", serveIMDSIssuerURL)
	}
	if serveIMDSPort < 1 || serveIMDSPort > 65535 {
		return serveIMDSConfig{}, fmt.Errorf("--port out of range: %d", serveIMDSPort)
	}

	return serveIMDSConfig{
		clientID:         serveIMDSClientID,
		clientSecretFile: serveIMDSClientSecretFile,
		issuerURL:        strings.TrimRight(serveIMDSIssuerURL, "/"),
		awsRoleARN:       serveIMDSAWSRoleARN,
		scope:            serveIMDSScope,
		awsRegion:        serveIMDSAWSRegion,
		bindAddress:      serveIMDSBindAddress,
		port:             serveIMDSPort,
		refreshBefore:    serveIMDSRefreshBefore,
		sessionDuration:  serveIMDSSessionDuration,
		requireIMDSv2:    serveIMDSRequireV2,
	}, nil
}

// loadServeIMDSEnv populates flag-backing variables from AWS_OIDC_IMDS_*
// env vars where the corresponding flag was not set on the command line.
// Flag value takes precedence over env var.
func loadServeIMDSEnv() {
	type binding struct {
		ptr    *string
		envKey string
	}
	for _, b := range []binding{
		{&serveIMDSClientID, "AWS_OIDC_IMDS_CLIENT_ID"},
		{&serveIMDSClientSecretFile, "AWS_OIDC_IMDS_CLIENT_SECRET_FILE"},
		{&serveIMDSIssuerURL, "AWS_OIDC_IMDS_ISSUER_URL"},
		{&serveIMDSAWSRoleARN, "AWS_OIDC_IMDS_AWS_ROLE_ARN"},
		{&serveIMDSScope, "AWS_OIDC_IMDS_SCOPE"},
		{&serveIMDSAWSRegion, "AWS_OIDC_IMDS_AWS_REGION"},
		{&serveIMDSBindAddress, "AWS_OIDC_IMDS_BIND_ADDRESS"},
	} {
		if *b.ptr == "" || (b.ptr == &serveIMDSAWSRegion && *b.ptr == "us-west-2") || (b.ptr == &serveIMDSBindAddress && *b.ptr == defaultBindAddress) {
			if v := os.Getenv(b.envKey); v != "" {
				*b.ptr = v
			}
		}
	}
}

// readClientSecretFile reads the secret from disk, requires the file to be
// mode 0600 (warns if looser, fails if missing/unreadable), and returns the
// trimmed content. The secret is never written to logs.
func readClientSecretFile(p string) (string, error) {
	if p == "" {
		return "", errors.New("--client-secret-file is empty")
	}
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("cannot stat client_secret file %q: %w", p, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("client_secret file %q is a directory", p)
	}
	mode := info.Mode().Perm()
	if mode != 0o600 {
		slog.Warn("imds_server: client_secret file mode is not 0600",
			slog.String("path", p),
			slog.String("mode", fmt.Sprintf("%04o", mode)),
			slog.String("recommended", "0600"),
		)
	}
	contents, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("reading client_secret file %q: %w", p, err)
	}
	secret := strings.TrimSpace(string(contents))
	if secret == "" {
		return "", fmt.Errorf("client_secret file %q is empty", p)
	}
	return secret, nil
}

// roleNameFromARN extracts the role name from an IAM role ARN. The IMDS
// security-credentials path uses the role name (not the full ARN).
func roleNameFromARN(roleARN string) (string, error) {
	const prefix = "arn:aws:iam::"
	if !strings.HasPrefix(roleARN, prefix) {
		return "", fmt.Errorf("invalid role ARN %q: must start with %s", roleARN, prefix)
	}
	rest := roleARN[len(prefix):]
	// rest is "<account-id>:role/<name>" possibly with a path: "<account-id>:role/path/to/<name>"
	idx := strings.Index(rest, ":role/")
	if idx < 0 {
		return "", fmt.Errorf("invalid role ARN %q: missing :role/ segment", roleARN)
	}
	rolePath := rest[idx+len(":role/"):]
	name := path.Base(rolePath)
	if name == "" || name == "." || name == "/" {
		return "", fmt.Errorf("invalid role ARN %q: empty role name", roleARN)
	}
	return name, nil
}

// isLoopbackBindAddress returns true if the bind address is a loopback
// address (127.0.0.0/8, ::1, "localhost", or empty).
func isLoopbackBindAddress(addr string) bool {
	if addr == "" || addr == "localhost" {
		return true
	}
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}

// Reference aws to satisfy import; used when constructing aws.Config indirectly.
var _ = aws.AnonymousCredentials{}
