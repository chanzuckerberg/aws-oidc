package getter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/go-misc/oidc/v5/cli/client"
)

func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("creating AWS session: %w", err)
	}

	roleSessionName := token.Claims.Email
	if roleSessionName == "" {
		roleSessionName = token.Claims.PreferredUsername
		if roleSessionName == "" {
			roleSessionName = "unknown-username"
		}
	}

	stsSession := sts.New(sess)
	sessionDurationSeconds := int64(sessionDuration.Seconds())
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  &roleSessionName,
		WebIdentityToken: aws.String(token.IDToken),
		DurationSeconds:  &sessionDurationSeconds,
	}
	tokenOutput, err := stsSession.AssumeRoleWithWebIdentityWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("assuming role with web identity: %w", err)
	}
	if tokenOutput == nil {
		return nil, fmt.Errorf("nil Token from AWS")
	}
	return tokenOutput, nil
}
