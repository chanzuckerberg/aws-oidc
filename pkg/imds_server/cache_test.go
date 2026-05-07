package imds_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
)

// stsResponseXML returns a valid AssumeRoleWithWebIdentity XML response with
// the given expiry. The SDK parses this format by default.
func stsResponseXML(accessKey, secretKey, sessionToken string, expiry time.Time) string {
	return fmt.Sprintf(`<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleWithWebIdentityResult>
    <SubjectFromWebIdentityToken>0oa1234567890</SubjectFromWebIdentityToken>
    <Audience>https://test-audience</Audience>
    <AssumedRoleUser>
      <Arn>arn:aws:sts::123456789012:assumed-role/TestRole/test-session</Arn>
      <AssumedRoleId>AROATEST:test-session</AssumedRoleId>
    </AssumedRoleUser>
    <Provider>example.okta.com/oauth2/aus123</Provider>
    <Credentials>
      <AccessKeyId>%s</AccessKeyId>
      <SecretAccessKey>%s</SecretAccessKey>
      <SessionToken>%s</SessionToken>
      <Expiration>%s</Expiration>
    </Credentials>
  </AssumeRoleWithWebIdentityResult>
  <ResponseMetadata>
    <RequestId>req-1</RequestId>
  </ResponseMetadata>
</AssumeRoleWithWebIdentityResponse>`, accessKey, secretKey, sessionToken, expiry.UTC().Format(time.RFC3339))
}

// newMockSTS returns an httptest.Server that responds to
// AssumeRoleWithWebIdentity with the given expiry, and an atomic counter.
func newMockSTS(t *testing.T, expiry time.Time) (*httptest.Server, *atomic.Int32, *sts.Client) {
	t.Helper()
	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(stsResponseXML(
			fmt.Sprintf("AKIA%d", calls.Load()),
			fmt.Sprintf("secret-%d", calls.Load()),
			fmt.Sprintf("token-%d", calls.Load()),
			expiry,
		)))
	}))
	t.Cleanup(server.Close)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	require.NoError(t, err)
	stsClient := sts.NewFromConfig(cfg, func(o *sts.Options) {
		o.BaseEndpoint = aws.String(server.URL)
	})

	return server, &calls, stsClient
}

func TestCredentialsCache_Get_FirstCallFetches(t *testing.T) {
	r := require.New(t)
	expiry := time.Now().Add(1 * time.Hour)
	_, calls, stsClient := newMockSTS(t, expiry)

	var fetcherCalls atomic.Int32
	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			fetcherCalls.Add(1)
			return "fake.jwt", time.Now().Add(time.Hour), nil
		},
	})
	r.NoError(err)
	r.False(cache.Ready())

	creds, err := cache.Get(context.Background())
	r.NoError(err)
	r.Equal("AKIA1", creds.AccessKeyID)
	r.True(cache.Ready())
	r.Equal(int32(1), calls.Load())
	r.Equal(int32(1), fetcherCalls.Load())
}

func TestCredentialsCache_Get_SecondCallCached(t *testing.T) {
	r := require.New(t)
	expiry := time.Now().Add(1 * time.Hour) // well above the 5m refresh window
	_, calls, stsClient := newMockSTS(t, expiry)

	var fetcherCalls atomic.Int32
	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			fetcherCalls.Add(1)
			return "fake.jwt", time.Now().Add(time.Hour), nil
		},
	})
	r.NoError(err)

	for range 5 {
		_, err := cache.Get(context.Background())
		r.NoError(err)
	}
	// Only the first call should hit STS.
	r.Equal(int32(1), calls.Load(), "expected exactly 1 STS call across 5 Get calls")
	r.Equal(int32(1), fetcherCalls.Load())
}

func TestCredentialsCache_Get_RefreshAfterExpiry(t *testing.T) {
	r := require.New(t)
	// Expiry is in the past so every call refreshes.
	expiry := time.Now().Add(-1 * time.Minute)
	_, calls, stsClient := newMockSTS(t, expiry)

	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			return "fake.jwt", time.Now(), nil
		},
		// Tiny refresh window so we can reason about it.
		RefreshBefore: 1 * time.Second,
	})
	r.NoError(err)

	for range 3 {
		_, err := cache.Get(context.Background())
		r.NoError(err)
	}
	// Each Get sees expired creds and refreshes.
	r.Equal(int32(3), calls.Load(), "every call should refresh because creds are expired")
}

func TestCredentialsCache_Get_ConcurrentCoalesced(t *testing.T) {
	r := require.New(t)
	// Slow STS so multiple concurrent calls overlap.
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(stsResponseXML("AKIATEST", "secret", "token", time.Now().Add(1*time.Hour))))
	}))
	t.Cleanup(server.Close)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	r.NoError(err)
	stsClient := sts.NewFromConfig(cfg, func(o *sts.Options) {
		o.BaseEndpoint = aws.String(server.URL)
	})

	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			return "fake.jwt", time.Now().Add(time.Hour), nil
		},
	})
	r.NoError(err)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cache.Get(context.Background())
			r.NoError(err)
		}()
	}
	wg.Wait()
	// aws.CredentialsCache coalesces concurrent calls — exactly 1 STS call.
	r.Equal(int32(1), calls.Load(), "concurrent Get calls should be coalesced into a single STS call")
}

func TestCredentialsCache_Get_FetcherError(t *testing.T) {
	r := require.New(t)
	_, _, stsClient := newMockSTS(t, time.Now().Add(time.Hour))

	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			return "", time.Time{}, errors.New("okta unreachable")
		},
	})
	r.NoError(err)

	_, err = cache.Get(context.Background())
	r.Error(err)
	r.Contains(err.Error(), "okta unreachable")
	r.False(cache.Ready())
}

func TestCredentialsCache_Get_EmptyJWTError(t *testing.T) {
	r := require.New(t)
	_, _, stsClient := newMockSTS(t, time.Now().Add(time.Hour))

	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		Fetcher: func(ctx context.Context) (string, time.Time, error) {
			return "", time.Time{}, nil
		},
	})
	r.NoError(err)

	_, err = cache.Get(context.Background())
	r.Error(err)
	r.Contains(err.Error(), "empty JWT")
}

func TestNewCredentialsCache_RequiredFields(t *testing.T) {
	r := require.New(t)
	stsClient := &sts.Client{}

	_, err := NewCredentialsCache(NewCredentialsCacheOptions{})
	r.Error(err)
	r.Contains(err.Error(), "STSClient")

	_, err = NewCredentialsCache(NewCredentialsCacheOptions{STSClient: stsClient})
	r.Error(err)
	r.Contains(err.Error(), "RoleARN")

	_, err = NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient: stsClient,
		RoleARN:   "arn:aws:iam::123456789012:role/X",
	})
	r.Error(err)
	r.Contains(err.Error(), "Fetcher")
}
