package aws_config_client

import (
	"bytes"
	"fmt"
	"testing"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestSurveyProfiles(t *testing.T) {
	r := require.New(t)

	// note how: "Account Name With Spaces" => "account-name-with-spaces"
	expected := `[profile account-name-with-spaces]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile my-second-new-profile]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile test1]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test2RoleName 2> /dev/tty'
region             = test-region

`

	out := ini.Empty()

	// we add a junk section and make sure it disappears in the output
	junkSection, err := out.NewSection("profile test1")
	r.NoError(err)
	junkSection.Key("role_arn").SetValue("this should disappear")

	prompt := &MockPrompt{

		selectResponse: []int{
			1,    // select the profile method of configuring
			0, 0, // select the first role in the first account
			1, 0, // select the first role in the second account
			1, 1, // select the second role in the second account
		},
		inputResponse: []string{
			"test-region",                   // aws region
			"", "my-second-new-profile", "", // aws profile names
		},
		confirmResponse: []bool{true, true, false},
	}

	c := NewCompleter(prompt, generateDummyData())

	err = c.Complete(out)
	r.NoError(err)

	generatedConfig := bytes.NewBuffer(nil)
	_, err = out.WriteTo(generatedConfig)
	r.NoError(err)
	r.Equal(expected, generatedConfig.String())
}

func TestSurveyRoles(t *testing.T) {
	r := require.New(t)

	expected := `[profile account-name-with-spaces-test1RoleName]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile account-name-with-spaces]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile test1-test1RoleName]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile test1]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile test1-test2RoleName]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test2RoleName 2> /dev/tty'
region             = test-region

`
	out := ini.Empty()

	prompt := &MockPrompt{

		selectResponse: []int{
			0, // select the role method of configuring
			0, // select the first role
		},
		inputResponse: []string{
			"test-region", // aws region
		},
		confirmResponse: []bool{false},
	}

	c := NewCompleter(prompt, generateDummyData())

	err := c.Complete(out)
	r.NoError(err)

	generatedConfig := bytes.NewBuffer(nil)
	_, err = out.WriteTo(generatedConfig)
	r.NoError(err)
	r.Equal(expected, generatedConfig.String())
}

func TestNoRoles(t *testing.T) {
	r := require.New(t)
	expected := ``

	out := ini.Empty()
	prompt := &MockPrompt{
		selectResponse:  []int{},
		inputResponse:   []string{},
		confirmResponse: []bool{},
	}

	c := NewCompleter(prompt, generateEmptyData())

	err := c.Complete(out)
	r.NoError(err)

	generatedConfig := bytes.NewBuffer(nil)
	_, err = out.WriteTo(generatedConfig)
	r.NoError(err)
	r.Equal(expected, generatedConfig.String())
}

func TestAWSProfileNameValidator(t *testing.T) {
	type test struct {
		input interface{}
		err   error
	}
	r := require.New(t)

	tests := []test{
		{input: 1, err: fmt.Errorf("input not a string")},
		{input: "not valid", err: fmt.Errorf("Input (not valid) not a valid AWS profile name")},
		{input: "valid", err: nil},
	}

	c := NewCompleter(nil, generateDummyData())
	for _, test := range tests {
		err := c.awsProfileNameValidator(test.input)
		if test.err == nil {
			r.NoError(err)
		} else {
			r.Error(err)
			r.Equal(test.err.Error(), err.Error())
		}

	}
}

func TestCalCulateDefaultProfileName(t *testing.T) {
	type test struct {
		input server.AWSAccount
		output string
	}

	tests := []test{
		{
			input: server.AWSAccount{
				Name: "test1",
				ID:   "test_id_1",
				Alias: "",
			},
			output: "test1",
		},
		{
			input: server.AWSAccount{
				Name: "test2",
				ID:   "test_id_2",
				Alias: "alias2",
			},
			output: "alias2",
		},
	}

	r := require.New(t)

	c := NewCompleter(nil, generateDummyData())
	for _, test := range tests {
		profleName := c.calculateDefaultProfileName(test.input)
		r.Equal(test.output, profleName)
	}
}

func generateDummyData() *server.AWSConfig {
	return &server.AWSConfig{
		Profiles: []server.AWSProfile{
			{
				ClientID: "bar_client_id",
				AWSAccount: server.AWSAccount{
					Name: "test1",
					ID:   "test_id_1",
					Alias: "alias1",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test1RoleName",
			},
			{
				ClientID: "bar_client_id",
				AWSAccount: server.AWSAccount{
					Name: "test1",
					ID:   "test_id_1",
					Alias: "alias1",
				},
				RoleARN:   "test2RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test2RoleName",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name: "Account Name With Spaces",
					ID:   "account id 2",
					Alias: "alias2",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test1RoleName",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name: "Account Name With Spaces",
					ID:   "account id 2",
					Alias: "alias2",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test1RoleName",
			},
		},
	}
}

func generateEmptyData() *server.AWSConfig {
	return &server.AWSConfig{
		Profiles: []server.AWSProfile{},
	}
}
