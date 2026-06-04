package getter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
)

func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	// AssumeRoleWithWebIdentity is a global STS operation; the region only
	// selects the endpoint. A credential_process subprocess inherits neither the
	// caller's --region flag nor the profile's region, and the SDK endpoint
	// resolver rejects an empty region ("Missing Region"), so fall back to
	// us-east-1 when nothing else resolves one. An explicit AWS_REGION /
	// AWS_DEFAULT_REGION still takes precedence over this default.
	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	roleSessionName := token.Claims.Email
	if roleSessionName == "" {
		roleSessionName = token.Claims.PreferredUsername
		if roleSessionName == "" {
			roleSessionName = "unknown-username"
		}
	}

	stsClient := sts.NewFromConfig(cfg)
	sessionDurationSeconds := int32(sessionDuration.Seconds())
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  &roleSessionName,
		WebIdentityToken: aws.String(token.IDToken),
		DurationSeconds:  &sessionDurationSeconds,
	}
	tokenOutput, err := stsClient.AssumeRoleWithWebIdentity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("assuming role with web identity: %w", err)
	}
	if tokenOutput == nil {
		return nil, fmt.Errorf("nil Token from AWS")
	}
	return tokenOutput, nil
}
