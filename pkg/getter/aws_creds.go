package getter

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
)

// stsSessionNameInvalid matches characters NOT permitted in the STS
// RoleSessionName. The valid set is alphanumeric plus =,.@_-
var stsSessionNameInvalid = regexp.MustCompile(`[^A-Za-z0-9=,.@_-]`)

// stsSessionNameMaxLen is the AWS STS RoleSessionName length cap.
const stsSessionNameMaxLen = 64

// GetAWSAssumeIdentity exchanges a chanzuckerberg/go-misc OIDC token for an
// AWS STS session by calling sts:AssumeRoleWithWebIdentity. This is the
// human-flow path; the role session name is derived from the token's email
// (or preferredUsername) claim.
//
// For workload (non-human) flows, prefer AssumeRoleWithJWT, which takes a
// raw JWT and an explicit session name.
func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	sessionName := token.Claims.Email
	if sessionName == "" {
		sessionName = token.Claims.PreferredUsername
		if sessionName == "" {
			sessionName = "unknown-username"
		}
	}
	return AssumeRoleWithJWT(ctx, token.IDToken, sessionName, roleARN, sessionDuration)
}

// AssumeRoleWithJWT calls sts:AssumeRoleWithWebIdentity with the given JWT.
// The session name is sanitized to satisfy the STS RoleSessionName regex
// ([A-Za-z0-9=,.@_-]+), truncated to 64 characters. AWS region and other
// SDK configuration are read from the default credential chain (env, shared
// config); callers may pass optional optFns to override (useful for tests
// that need a custom BaseEndpoint).
//
// This is the workload-flow entry point: for example, a JWT obtained via
// Okta's OAuth 2.0 client_credentials grant can be exchanged here for STS
// credentials usable to sign AWS API requests.
func AssumeRoleWithJWT(
	ctx context.Context,
	jwt string,
	sessionName string,
	roleARN string,
	sessionDuration time.Duration,
	optFns ...func(*config.LoadOptions) error,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	if jwt == "" {
		return nil, fmt.Errorf("AssumeRoleWithJWT: empty JWT")
	}
	if roleARN == "" {
		return nil, fmt.Errorf("AssumeRoleWithJWT: empty roleARN")
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	cleanedSessionName := sanitizeSessionName(sessionName)

	stsClient := sts.NewFromConfig(cfg)
	durationSeconds := int32(sessionDuration.Seconds())
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  aws.String(cleanedSessionName),
		WebIdentityToken: aws.String(jwt),
		DurationSeconds:  &durationSeconds,
	}
	out, err := stsClient.AssumeRoleWithWebIdentity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("assuming role with web identity: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("nil token from AWS STS")
	}
	return out, nil
}

// sanitizeSessionName returns a string that satisfies the STS RoleSessionName
// regex `[A-Za-z0-9=,.@_-]+` and length cap of 64. Disallowed characters are
// replaced with `-`. Leading/trailing `-` are stripped. Empty input yields
// "session" so the caller never has to pass a guaranteed-valid string.
func sanitizeSessionName(s string) string {
	if s == "" {
		return "session"
	}
	cleaned := stsSessionNameInvalid.ReplaceAllString(s, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		return "session"
	}
	if len(cleaned) > stsSessionNameMaxLen {
		cleaned = cleaned[:stsSessionNameMaxLen]
	}
	return cleaned
}
