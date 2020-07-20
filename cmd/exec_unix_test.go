package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestAWSEnvPrecedence(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())

	config := ini.Empty()
	acct1Section, err := config.NewSection("profile Account1")
	r.NoError(err)
	prevRegionKey1, err := acct1Section.NewKey("region", "dummyregion1")
	r.NoError(err)
	prevOutputKey1, err := acct1Section.NewKey("output", "dummyoutput1")
	r.NoError(err)

	// Next steps: load the profile, load environment variables, see that the environment variables were set
	err = os.Setenv("AWS_OUTPUT", "dummyoutput2")
	r.NoError(err)
	err = os.Setenv("AWS_REGION", "dummyregion2")
	r.NoError(err)

	// If there aren't any environment variables, make sure that the region/output is set anyway
	awsOIDCConfig := &aws_config_client.AWSOIDCConfiguration{
		ClientID:  "dummyClientID",
		IssuerURL: "localhost",
		RoleARN:   "dummyRoleARN",
		Region:    aws.String(prevRegionKey1.String()),
		Output:    aws.String(prevOutputKey1.String()),
	}
	dummyCredentials := &sts.Credentials{
		AccessKeyId:     aws.String("dummyAccessKeyId"),
		SecretAccessKey: aws.String("dummySecretAccessKey"),
		SessionToken:    aws.String("dummySessionToken"),
	}

	// Get the AWS Environment variables
	envVars := getAWSEnvVars(&sts.AssumeRoleWithWebIdentityOutput{Credentials: dummyCredentials}, awsOIDCConfig)
	// Running an easy command
	err = exec(context.Background(), "ls", nil, envVars)
	r.NoError(err)

	regionVal, present := os.LookupEnv("AWS_REGION")
	r.True(present)
	r.Equal(regionVal, "dummyregion2")
	r.NotEqual(regionVal, "dummyregion1")

	outputVal, present := os.LookupEnv("AWS_OUTPUT")
	r.True(present)
	r.Equal(outputVal, "dummyoutput2")
	r.NotEqual(outputVal, "dummyoutput1")

	// Remove the AWS_OUTPUT and AWS_REGION values
	err = os.Unsetenv("AWS_OUTPUT")
	r.NoError(err)
	err = os.Unsetenv("AWS_REGION")
	r.Error(err)
}
