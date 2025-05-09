package getter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	client "github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/client"
)

func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Unable to create AWS session: %w", err)
	}

	stsSession := sts.New(sess)
	sessionDurationSeconds := int64(sessionDuration.Seconds())
	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  aws.String(token.Claims.Email),
		WebIdentityToken: aws.String(token.IDToken),
		DurationSeconds:  &sessionDurationSeconds,
	}
	tokenOutput, err := stsSession.AssumeRoleWithWebIdentityWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("AWS AssumeRoleWithWebIdentity error: %w", err)
	}
	if tokenOutput == nil {
		return nil, fmt.Errorf("nil Token from AWS")
	}
	return tokenOutput, nil
}
