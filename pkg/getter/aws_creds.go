package getter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
)

func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	ctx, span := beeline.StartSpan(ctx, "assume_role_with_web_identity")
	defer span.Send()

	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create AWS session")
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
		return nil, errors.Wrap(err, "AWS AssumeRoleWithWebIdentity error")
	}
	if tokenOutput == nil {
		return nil, errors.New("nil Token from AWS")
	}
	return tokenOutput, nil
}
