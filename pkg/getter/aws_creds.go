package getter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	client "github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/pkg/errors"
)

func GetAWSAssumeIdentity(
	ctx context.Context,
	token *client.Token,
	roleARN string,
	sessionDuration time.Duration,
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	sess, err := session.NewSession()
	if err != nil {
		server.AddBeelineFields(ctx, server.BeelineField{
			Key:   "AWS Session error",
			Value: err.Error(),
		})
		return nil, errors.Wrap(err, "Unable to create AWS session")
	}

	stsSession := sts.New(sess)
	sessionDurationSeconds := int64(sessionDuration.Seconds())
	server.AddBeelineFields(ctx, server.BeelineField{
		Key:   "AWS STS Session Duration",
		Value: sessionDuration.String(),
	})

	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  aws.String(token.Claims.Email),
		WebIdentityToken: aws.String(token.IDToken),
		DurationSeconds:  &sessionDurationSeconds,
	}
	tokenOutput, err := stsSession.AssumeRoleWithWebIdentityWithContext(ctx, input)
	if err != nil {
		server.AddBeelineFields(ctx, server.BeelineField{
			Key:   "AWS STS Session AssumeRoleWithWebIdentityWithContext error",
			Value: err.Error(),
		})
		return nil, errors.Wrap(err, "AWS AssumeRoleWithWebIdentity error")
	}
	if tokenOutput == nil {
		return nil, errors.New("nil Token from AWS")
	}
	return tokenOutput, nil
}
