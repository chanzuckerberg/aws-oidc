package imds_server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/stretchr/testify/require"
)

// dummyCache is a CredentialsCache stand-in for tests that need to control
// the credential value or simulate an error.
func dummyCache(t *testing.T, creds aws.Credentials, retrieveErr error) *CredentialsCache {
	t.Helper()
	return newTestCache(creds, retrieveErr)
}

// newTestCache builds a CredentialsCache that returns the given credentials
// or error on Get. It bypasses the real STS provider.
func newTestCache(creds aws.Credentials, retrieveErr error) *CredentialsCache {
	c := &CredentialsCache{}
	c.inner = stubProvider{creds: creds, err: retrieveErr}
	return c
}

type stubProvider struct {
	creds aws.Credentials
	err   error
}

func (s stubProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return s.creds, s.err
}

func startServer(t *testing.T, opts Options) *httptest.Server {
	t.Helper()
	srv, err := NewServer(opts)
	require.NoError(t, err)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func TestServer_RequiresRoleNameAndCache(t *testing.T) {
	r := require.New(t)
	_, err := NewServer(Options{})
	r.Error(err)
	r.Contains(err.Error(), "RoleName")

	_, err = NewServer(Options{RoleName: "x"})
	r.Error(err)
	r.Contains(err.Error(), "Cache")
}

func TestPUT_TokenIncludesTTLHeader(t *testing.T) {
	r := require.New(t)
	creds := aws.Credentials{
		AccessKeyID:     "AKIA1",
		SecretAccessKey: "secret",
		SessionToken:    "token",
		Expires:         time.Now().Add(time.Hour),
		CanExpire:       true,
	}
	ts := startServer(t, Options{
		RoleName: "argus-amp-producer-test",
		Cache:    dummyCache(t, creds, nil),
	})

	req, err := http.NewRequest(http.MethodPut, ts.URL+tokenPath, nil)
	r.NoError(err)
	req.Header.Set(imdsTokenTTLHeader, "60")

	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)

	// Critical: the response MUST include the TTL header.
	r.NotEmpty(resp.Header.Get(imdsTokenTTLHeader),
		"PUT /latest/api/token must include X-Aws-Ec2-Metadata-Token-Ttl-Seconds response header; SDK errors without it")
	r.Equal("60", resp.Header.Get(imdsTokenTTLHeader))

	body, err := io.ReadAll(resp.Body)
	r.NoError(err)
	r.NotEmpty(body, "token body should be non-empty")
	r.Len(string(body), 64, "32-byte hex token should be 64 chars")
}

func TestPUT_TokenInvalidTTL(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, aws.Credentials{}, nil),
	})

	req, err := http.NewRequest(http.MethodPut, ts.URL+tokenPath, nil)
	r.NoError(err)
	req.Header.Set(imdsTokenTTLHeader, "abc")
	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestPUT_TokenWithoutTTL_UsesDefault(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, aws.Credentials{}, nil),
	})

	req, err := http.NewRequest(http.MethodPut, ts.URL+tokenPath, nil)
	r.NoError(err)
	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)
	r.Equal("21600", resp.Header.Get(imdsTokenTTLHeader))
}

func TestGET_RoleListAndCredentials_V2Flow(t *testing.T) {
	r := require.New(t)
	creds := aws.Credentials{
		AccessKeyID:     "AKIA-TEST",
		SecretAccessKey: "secret-test",
		SessionToken:    "token-test",
		Expires:         time.Now().Add(time.Hour),
		CanExpire:       true,
	}
	ts := startServer(t, Options{
		RoleName: "argus-test-role",
		Cache:    dummyCache(t, creds, nil),
	})

	// 1. PUT to mint a v2 session token.
	req, err := http.NewRequest(http.MethodPut, ts.URL+tokenPath, nil)
	r.NoError(err)
	req.Header.Set(imdsTokenTTLHeader, "300")
	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	tokenBody, err := io.ReadAll(resp.Body)
	r.NoError(err)
	resp.Body.Close()
	imdsTok := string(tokenBody)
	r.NotEmpty(imdsTok)

	// 2. GET role list with the token.
	req, err = http.NewRequest(http.MethodGet, ts.URL+roleListPath, nil)
	r.NoError(err)
	req.Header.Set(imdsTokenHeader, imdsTok)
	resp, err = http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	r.NoError(err)
	r.Equal("argus-test-role", string(body))

	// 3. GET credentials.
	req, err = http.NewRequest(http.MethodGet, ts.URL+roleCredsPathPrefix+"argus-test-role", nil)
	r.NoError(err)
	req.Header.Set(imdsTokenHeader, imdsTok)
	resp, err = http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	r.NoError(err)
	bodyStr := string(body)
	r.Contains(bodyStr, `"Code":"Success"`)
	r.Contains(bodyStr, `"AccessKeyId":"AKIA-TEST"`)
	r.Contains(bodyStr, `"SecretAccessKey":"secret-test"`)
	r.Contains(bodyStr, `"Token":"token-test"`)
	r.Contains(bodyStr, `"Expiration":"`)
	// Expiration must be RFC 3339 — the AWS SDK rejects other formats.
	expField := extractField(bodyStr, "Expiration")
	_, err = time.Parse(time.RFC3339, expField)
	r.NoError(err, "Expiration must parse as RFC 3339 (got %q)", expField)
}

func TestGET_Credentials_V1Flow_NoToken(t *testing.T) {
	r := require.New(t)
	creds := aws.Credentials{
		AccessKeyID:     "AKIA-V1",
		SecretAccessKey: "secret",
		SessionToken:    "token",
		Expires:         time.Now().Add(time.Hour),
		CanExpire:       true,
	}
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, creds, nil),
	})

	resp, err := http.Get(ts.URL + roleCredsPathPrefix + "argus")
	r.NoError(err)
	defer resp.Body.Close()
	// Default RequireIMDSv2=false: v1 path accepted.
	r.Equal(http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	r.Contains(string(body), `"Code":"Success"`)
}

func TestGET_Credentials_RequireIMDSv2_Refuses_V1(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName:      "argus",
		Cache:         dummyCache(t, aws.Credentials{}, nil),
		RequireIMDSv2: true,
	})

	resp, err := http.Get(ts.URL + roleCredsPathPrefix + "argus")
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestGET_Credentials_V2_InvalidToken_401(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, aws.Credentials{}, nil),
	})

	req, _ := http.NewRequest(http.MethodGet, ts.URL+roleCredsPathPrefix+"argus", nil)
	req.Header.Set(imdsTokenHeader, "not-a-real-token")
	resp, err := http.DefaultClient.Do(req)
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestGET_Credentials_RefreshFailure_500(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, aws.Credentials{}, errors.New("okta unreachable")),
	})

	resp, err := http.Get(ts.URL + roleCredsPathPrefix + "argus")
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusInternalServerError, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	r.Contains(string(body), `"Code":"AssumeRoleUnauthorizedAccess"`)
	r.Contains(string(body), "okta unreachable")
}

func TestGET_UnknownRole_404(t *testing.T) {
	r := require.New(t)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    dummyCache(t, aws.Credentials{}, nil),
	})

	resp, err := http.Get(ts.URL + roleCredsPathPrefix + "wrong-role")
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestHealthAndReady(t *testing.T) {
	r := require.New(t)
	cache := dummyCache(t, aws.Credentials{
		AccessKeyID: "x", SecretAccessKey: "y", SessionToken: "z",
		Expires: time.Now().Add(time.Hour), CanExpire: true,
	}, nil)
	ts := startServer(t, Options{
		RoleName: "argus",
		Cache:    cache,
	})

	// /healthz always 200
	resp, err := http.Get(ts.URL + healthzPath)
	r.NoError(err)
	resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)

	// /readyz starts unready
	resp, err = http.Get(ts.URL + readyzPath)
	r.NoError(err)
	resp.Body.Close()
	r.Equal(http.StatusServiceUnavailable, resp.StatusCode)

	// After one successful Get, ready
	_, err = cache.Get(context.Background())
	r.NoError(err)
	resp, err = http.Get(ts.URL + readyzPath)
	r.NoError(err)
	resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)
}

// TestSDKCompatibility_V2 verifies the AWS SDK Go v2's IMDS provider can
// successfully fetch credentials from our server using the v2 token dance.
// This is the most important regression test: any wire-shape mistake here
// breaks the whole point of the helper.
func TestSDKCompatibility_V2(t *testing.T) {
	r := require.New(t)
	want := aws.Credentials{
		AccessKeyID:     "AKIA-SDKTEST",
		SecretAccessKey: "secret-sdktest",
		SessionToken:    "token-sdktest",
		Expires:         time.Now().Add(30 * time.Minute).UTC(),
		CanExpire:       true,
	}
	ts := startServer(t, Options{
		RoleName: "argus-sdktest",
		Cache:    dummyCache(t, want, nil),
	})

	// Build an IMDS client pointed at our server.
	client := imds.New(imds.Options{
		Endpoint: ts.URL,
	})

	// First, the SDK fetches the role name…
	roleOut, err := client.GetMetadata(context.Background(), &imds.GetMetadataInput{
		Path: "iam/security-credentials/",
	})
	r.NoError(err)
	defer roleOut.Content.Close()
	roleNameBytes, err := io.ReadAll(roleOut.Content)
	r.NoError(err)
	r.Equal("argus-sdktest", string(roleNameBytes))

	// …then per-role credentials.
	credsOut, err := client.GetMetadata(context.Background(), &imds.GetMetadataInput{
		Path: "iam/security-credentials/argus-sdktest",
	})
	r.NoError(err)
	defer credsOut.Content.Close()
	credsBody, err := io.ReadAll(credsOut.Content)
	r.NoError(err)
	r.Contains(string(credsBody), `"Code":"Success"`)
	r.Contains(string(credsBody), `"AccessKeyId":"AKIA-SDKTEST"`)
}

// TestSDKCompatibility_V1Fallback verifies that when our server doesn't require
// v2, the v1 path (no token header) still works for the SDK.
func TestSDKCompatibility_V1Fallback(t *testing.T) {
	r := require.New(t)
	want := aws.Credentials{
		AccessKeyID:     "AKIA-V1",
		SecretAccessKey: "secret",
		SessionToken:    "token",
		Expires:         time.Now().Add(30 * time.Minute).UTC(),
		CanExpire:       true,
	}
	ts := startServer(t, Options{
		RoleName: "argus-v1",
		Cache:    dummyCache(t, want, nil),
	})

	// We test the v1 path manually since the SDK starts with v2.
	resp, err := http.Get(ts.URL + roleCredsPathPrefix + "argus-v1")
	r.NoError(err)
	defer resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)
}

// TestRedactSecretsInError verifies the JWT redaction helper.
func TestRedactSecretsInError(t *testing.T) {
	cases := map[string]struct {
		in       string
		notWant  string
		want     string
	}{
		"plain string": {
			in:   "okta unreachable",
			want: "okta unreachable",
		},
		"jwt is redacted": {
			in:      "got token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIwb2Ex.signaturepart_with_enough_chars and other text",
			notWant: "eyJhbGciOiJIUzI1NiJ9",
			want:    "[REDACTED-JWT]",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			got := redactSecretsInError(errors.New(tc.in))
			if tc.notWant != "" {
				r.NotContains(got, tc.notWant)
			}
			r.Contains(got, tc.want)
		})
	}
}

// extractField pulls a string-typed JSON field value out of a flat JSON
// object body. Best-effort; sufficient for asserting on Expiration etc.
func extractField(body, field string) string {
	prefix := `"` + field + `":"`
	i := strings.Index(body, prefix)
	if i < 0 {
		return ""
	}
	rest := body[i+len(prefix):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return ""
	}
	return rest[:end]
}
