package imds_server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/workload_oidc"
	"github.com/stretchr/testify/require"
)

// TestIntegration_FullPipeline wires the entire stack end-to-end:
//
//	Okta (httptest) -> workload_oidc -> stscreds -> STS (httptest)
//	-> CredentialsCache -> imds_server -> AWS SDK IMDS client
//
// Verifies that the AWS SDK's IMDS provider can fetch credentials from our
// server and that those credentials match what the STS stub vended.
func TestIntegration_FullPipeline(t *testing.T) {
	r := require.New(t)

	// 1. Mock Okta /v1/token endpoint.
	var oktaCalls atomic.Int32
	oktaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.Equal("/v1/token", req.URL.Path)
		r.NoError(req.ParseForm())
		r.Equal("client_credentials", req.Form.Get("grant_type"))
		r.Equal("test-client", req.Form.Get("client_id"))
		r.Equal("test-secret", req.Form.Get("client_secret"))
		// We must NOT send audience.
		r.Empty(req.Form.Get("audience"))

		oktaCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "fake.jwt.token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer oktaServer.Close()

	// 2. Mock AWS STS AssumeRoleWithWebIdentity.
	var stsCalls atomic.Int32
	stsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		stsCalls.Add(1)
		// Verify STS sees the JWT we got from Okta.
		r.NoError(req.ParseForm())
		r.Equal("fake.jwt.token", req.Form.Get("WebIdentityToken"))
		r.Equal("arn:aws:iam::123456789012:role/argus-amp-producer-test", req.Form.Get("RoleArn"))

		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(stsResponseXML(
			fmt.Sprintf("AKIA-INTEGRATION-%d", stsCalls.Load()),
			fmt.Sprintf("secret-integration-%d", stsCalls.Load()),
			fmt.Sprintf("token-integration-%d", stsCalls.Load()),
			time.Now().Add(time.Hour).UTC(),
		)))
	}))
	defer stsServer.Close()

	// 3. Build the workload_oidc Config pointed at the mock Okta.
	oidcCfg := workload_oidc.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		IssuerURL:    oktaServer.URL,
	}

	// 4. Build an STS client pointed at the mock STS.
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-west-2"),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	r.NoError(err)
	stsClient := sts.NewFromConfig(awsCfg, func(o *sts.Options) {
		o.BaseEndpoint = aws.String(stsServer.URL)
	})

	// 5. Wire the cache.
	cache, err := NewCredentialsCache(NewCredentialsCacheOptions{
		STSClient:       stsClient,
		RoleARN:         "arn:aws:iam::123456789012:role/argus-amp-producer-test",
		RoleSessionName: "test-client",
		SessionDuration: time.Hour,
		Fetcher:         oidcCfg.FetchToken,
	})
	r.NoError(err)

	// 6. Spin up the IMDS server in front of a real http.Server.
	srv, err := NewServer(Options{
		RoleName: "argus-amp-producer-test",
		Cache:    cache,
	})
	r.NoError(err)

	imdsHTTPServer := httptest.NewServer(srv.Handler())
	defer imdsHTTPServer.Close()

	// 7. Use the real AWS SDK IMDS client to fetch credentials from our server.
	imdsClient := imds.New(imds.Options{Endpoint: imdsHTTPServer.URL})

	// First call: triggers full Okta -> STS round trip.
	roleOut, err := imdsClient.GetMetadata(context.Background(), &imds.GetMetadataInput{
		Path: "iam/security-credentials/",
	})
	r.NoError(err)
	roleNameBytes, _ := io.ReadAll(roleOut.Content)
	roleOut.Content.Close()
	r.Equal("argus-amp-producer-test", string(roleNameBytes))

	credsOut, err := imdsClient.GetMetadata(context.Background(), &imds.GetMetadataInput{
		Path: "iam/security-credentials/argus-amp-producer-test",
	})
	r.NoError(err)
	credsBody, _ := io.ReadAll(credsOut.Content)
	credsOut.Content.Close()
	r.Contains(string(credsBody), `"AccessKeyId":"AKIA-INTEGRATION-1"`)
	r.Contains(string(credsBody), `"Code":"Success"`)

	r.Equal(int32(1), oktaCalls.Load(), "exactly 1 Okta token call for first credential fetch")
	r.Equal(int32(1), stsCalls.Load(), "exactly 1 STS AssumeRole call for first credential fetch")

	// Second call should use cached credentials — no new Okta or STS calls.
	credsOut2, err := imdsClient.GetMetadata(context.Background(), &imds.GetMetadataInput{
		Path: "iam/security-credentials/argus-amp-producer-test",
	})
	r.NoError(err)
	_, _ = io.ReadAll(credsOut2.Content)
	credsOut2.Content.Close()

	r.Equal(int32(1), oktaCalls.Load(), "second IMDS call must hit cached creds (no Okta call)")
	r.Equal(int32(1), stsCalls.Load(), "second IMDS call must hit cached creds (no STS call)")
}

// TestIntegration_GracefulShutdown verifies that http.Server.Shutdown drains
// in-flight requests when the server context is cancelled.
func TestIntegration_GracefulShutdown(t *testing.T) {
	r := require.New(t)

	creds := aws.Credentials{
		AccessKeyID: "x", SecretAccessKey: "y", SessionToken: "z",
		Expires: time.Now().Add(time.Hour), CanExpire: true,
	}
	srv, err := NewServer(Options{
		RoleName: "argus",
		Cache:    newTestCache(creds, nil),
	})
	r.NoError(err)

	// Use a custom http.Server (not httptest) so we can call Shutdown.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	r.NoError(err)

	httpSrv := &http.Server{Handler: srv.Handler()}
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpSrv.Serve(listener)
	}()

	url := "http://" + listener.Addr().String()

	// Sanity: first request works.
	resp, err := http.Get(url + healthzPath)
	r.NoError(err)
	resp.Body.Close()
	r.Equal(http.StatusOK, resp.StatusCode)

	// Shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r.NoError(httpSrv.Shutdown(shutdownCtx))

	// Wait for Serve to return.
	select {
	case err := <-serveErr:
		// http.ErrServerClosed is the expected signal.
		r.ErrorIs(err, http.ErrServerClosed)
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return after Shutdown")
	}

	// Subsequent connections should be refused.
	_, err = http.Get(url + healthzPath)
	r.Error(err, "post-shutdown requests should fail")
}
