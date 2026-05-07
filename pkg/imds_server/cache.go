package imds_server

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// TokenFetcher is the function signature for obtaining a fresh JWT to
// exchange for STS credentials. workload_oidc.Config.FetchToken satisfies
// this signature.
type TokenFetcher func(ctx context.Context) (jwt string, expiresAt time.Time, err error)

// CredentialsCache wraps an aws.CredentialsCache fed by a
// stscreds.WebIdentityRoleProvider whose IdentityTokenRetriever calls a
// caller-provided TokenFetcher. The aws.CredentialsCache handles pre-expiry
// refresh and concurrent-caller coalescing internally.
type CredentialsCache struct {
	inner         aws.CredentialsProvider
	everSucceeded atomic.Bool
}

// NewCredentialsCacheOptions configures NewCredentialsCache. Required fields:
// STSClient, RoleARN, Fetcher.
type NewCredentialsCacheOptions struct {
	// STSClient is the AWS STS client used to call AssumeRoleWithWebIdentity.
	// Construct with sts.NewFromConfig(cfg). For tests, pass a client whose
	// config.BaseEndpoint points at a mock HTTP server.
	STSClient *sts.Client

	// RoleARN is the IAM role to assume.
	RoleARN string

	// RoleSessionName is the session name used in the AssumeRole call. The
	// caller is responsible for any sanitization (see getter.sanitizeSessionName).
	RoleSessionName string

	// SessionDuration is the requested STS DurationSeconds. Capped by the
	// IAM role's MaxSessionDuration; if SessionDuration > role cap, STS
	// rejects the request.
	SessionDuration time.Duration

	// Fetcher returns a fresh JWT to be exchanged for STS credentials.
	Fetcher TokenFetcher

	// RefreshBefore is the pre-expiry window in which a Retrieve call will
	// trigger a refresh. Default: 5 minutes.
	RefreshBefore time.Duration

	// TokenFetchTimeout bounds how long a single Fetcher call may take.
	// Default: 30 seconds.
	TokenFetchTimeout time.Duration
}

// NewCredentialsCache builds a CredentialsCache around
// stscreds.WebIdentityRoleProvider and aws.CredentialsCache.
func NewCredentialsCache(opts NewCredentialsCacheOptions) (*CredentialsCache, error) {
	if opts.STSClient == nil {
		return nil, errors.New("imds_server: NewCredentialsCache: STSClient is required")
	}
	if opts.RoleARN == "" {
		return nil, errors.New("imds_server: NewCredentialsCache: RoleARN is required")
	}
	if opts.Fetcher == nil {
		return nil, errors.New("imds_server: NewCredentialsCache: Fetcher is required")
	}

	fetchTimeout := opts.TokenFetchTimeout
	if fetchTimeout == 0 {
		fetchTimeout = 30 * time.Second
	}

	retriever := tokenRetrieverFn(func() ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		jwt, _, err := opts.Fetcher(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching identity token: %w", err)
		}
		if jwt == "" {
			return nil, errors.New("fetcher returned empty JWT")
		}
		return []byte(jwt), nil
	})

	provider := stscreds.NewWebIdentityRoleProvider(
		opts.STSClient,
		opts.RoleARN,
		retriever,
		func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = opts.RoleSessionName
			if opts.SessionDuration > 0 {
				o.Duration = opts.SessionDuration
			}
		},
	)

	refreshBefore := opts.RefreshBefore
	if refreshBefore == 0 {
		refreshBefore = 5 * time.Minute
	}

	cached := aws.NewCredentialsCache(provider, func(o *aws.CredentialsCacheOptions) {
		o.ExpiryWindow = refreshBefore
		o.ExpiryWindowJitterFrac = 0.2
	})

	return &CredentialsCache{inner: cached}, nil
}

// Get returns current credentials, refreshing via the underlying provider if
// the cached credentials are within the configured RefreshBefore window of
// expiry. Concurrent callers are coalesced by aws.CredentialsCache so a
// single refresh is shared across them.
func (c *CredentialsCache) Get(ctx context.Context) (aws.Credentials, error) {
	creds, err := c.inner.Retrieve(ctx)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("retrieving credentials: %w", err)
	}
	c.everSucceeded.Store(true)
	return creds, nil
}

// Ready reports whether the cache has produced at least one successful
// credential retrieval since startup. Used to drive readiness probes.
func (c *CredentialsCache) Ready() bool {
	return c.everSucceeded.Load()
}

// tokenRetrieverFn adapts a function to stscreds.IdentityTokenRetriever.
type tokenRetrieverFn func() ([]byte, error)

func (f tokenRetrieverFn) GetIdentityToken() ([]byte, error) { return f() }
