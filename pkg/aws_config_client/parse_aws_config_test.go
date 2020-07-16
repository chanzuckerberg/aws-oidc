package aws_config_client

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestResolveProfile(t *testing.T) {
	r := require.New(t)

	// https://golang.org/src/os/env_test.go
	defer util.ResetEnv(os.Environ())
	r.NoError(os.Unsetenv(envAWSProfile))

	// default
	prof, err := resolveProfile(nil)
	r.NoError(err)
	r.Equal(defaultAWSProfile, prof)

	// from env
	expectedProfile := "asdfasdfalkwq;e"
	os.Setenv(envAWSProfile, expectedProfile)
	prof, err = resolveProfile(nil)
	r.NoError(err)
	r.Equal(expectedProfile, prof)

	// flag
	var flagVal string

	expectedProfile = "flag-profile"
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(
		&flagVal,
		FlagProfile,
		"",
		"AWS Profile to fetch credentials from.")

	r.NoError(cmd.Flags().Set(
		FlagProfile,
		expectedProfile))

	prof, err = resolveProfile(cmd)
	r.NoError(err)
	r.Equal(expectedProfile, prof)
}

func TestExtractFromCredsProcess(t *testing.T) {
	type testData struct {
		in       string
		expected string
	}

	r := require.New(t)

	tests := []testData{
		{
			in:       "sh -c 'aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role'",
			expected: "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
		},
		{
			in:       "sh -c 'aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role 2> /dev/tty'",
			expected: "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
		},
		{
			in:       "sh -c \"aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role\"",
			expected: "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
		},
		{
			in:       "sh -c \"aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role 2> /dev/tty\"",
			expected: "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
		},
		{
			in:       "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
			expected: "aws-oidc creds-process --issuer-url=https://localhost --client-id=some-client-id --aws-role-arn=arn:aws:iam::12345678910:role/role",
		},
	}

	for _, test := range tests {
		out := cleanCredProcessCommand(test.in)
		r.Equal(test.expected, out)
	}
}
