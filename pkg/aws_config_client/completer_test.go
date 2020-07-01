package aws_config_client

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestRemoveOldProfile(t *testing.T) {
	r := require.New(t)
	baseAWSConfig := ini.Empty()
	// we add a junk section and make sure it disappears in the output
	junkSection, err := baseAWSConfig.NewSection("profile test1")
	r.NoError(err)
	junkSection.Key("output").SetValue("old_output")
	junkSection.Key("credential_process").SetValue("old_cred_process")
	junkSection.Key("region").SetValue("old_region")

	expected := `[profile test1]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

`
	prompt := &MockPrompt{

		selectResponse: []int{
			1,    // select the profile method of configuring
			1, 0, // select the first role in the first account
		},
		inputResponse: []string{
			"test-region", // aws region
			"",            // aws profile names
		},
		confirmResponse: []bool{false, true},
	}

	c := NewCompleter(prompt, generateDummyData())

	testWriter := bytes.NewBuffer(nil)
	err = c.Complete(baseAWSConfig, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestSurveyProfiles(t *testing.T) {
	r := require.New(t)

	// note how: "Account Name With Spaces" => "account-name-with-spaces"
	expected := `[profile test1]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile account-name-with-spaces]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

[profile my-second-new-profile]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName 2> /dev/tty'
region             = test-region

`

	baseAWSConfig := ini.Empty()

	prompt := &MockPrompt{

		selectResponse: []int{
			1,    // select the profile method of configuring
			1, 0, // select the first role in the test1 account
			0, 0, // select the first role in the account-name-with-spaces account
			1, 0, // select the first role in the test1 account so we can name it my-second-new-profile
		},
		inputResponse: []string{
			"test-region",                   // aws region
			"", "", "my-second-new-profile", // use default aws profile names
		},
		confirmResponse: []bool{true, true, false, true},
	}

	c := NewCompleter(prompt, generateDummyData())

	testWriter := bytes.NewBuffer(nil)
	err := c.Complete(baseAWSConfig, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
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

[profile account-name-with-spaces-test2RoleName]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test2RoleName 2> /dev/tty'
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
	newAWSProfiles := ini.Empty()

	prompt := &MockPrompt{

		selectResponse: []int{
			0, // select the role method of configuring
			0, // select the first role
		},
		inputResponse: []string{
			"test-region", // aws region
		},
		confirmResponse: []bool{true},
	}

	c := NewCompleter(prompt, generateDummyData())

	testWriter := bytes.NewBuffer(nil)
	err := c.Complete(newAWSProfiles, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestNoRoles(t *testing.T) {
	r := require.New(t)
	expected := ``

	newAWSProfiles := ini.Empty()
	prompt := &MockPrompt{
		selectResponse:  []int{},
		inputResponse:   []string{},
		confirmResponse: []bool{},
	}

	c := NewCompleter(prompt, generateEmptyData())

	testWriter, err := os.OpenFile("testfile", os.O_WRONLY|os.O_CREATE, 0600)
	defer testWriter.Close()
	r.NoError(err)
	err = c.Complete(newAWSProfiles, testWriter)
	r.NoError(err)

	generatedConfig := bytes.NewBuffer(nil)
	_, err = newAWSProfiles.WriteTo(generatedConfig)
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

func TestCalculateDefaultProfileName(t *testing.T) {
	type test struct {
		input  server.AWSAccount
		output string
	}

	tests := []test{
		{
			input: server.AWSAccount{
				Name:  "test1",
				ID:    "test_id_1",
				Alias: "",
			},
			output: "test1",
		},
		{
			input: server.AWSAccount{
				Name:  "test2",
				ID:    "test_id_2",
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
					Name:  "test1",
					ID:    "test_id_1",
					Alias: "test1",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test1RoleName",
			},
			{
				ClientID: "bar_client_id",
				AWSAccount: server.AWSAccount{
					Name:  "test1",
					ID:    "test_id_1",
					Alias: "test1",
				},
				RoleARN:   "test2RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test2RoleName",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name:  "Account Name With Spaces",
					ID:    "account id 2",
					Alias: "Account Name With Spaces",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test1RoleName",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name:  "Account Name With Spaces",
					ID:    "account id 2",
					Alias: "Account Name With Spaces",
				},
				RoleARN:   "test2RoleName",
				IssuerURL: "issuer-url",
				RoleName:  "test2RoleName",
			},
		},
	}
}

func generateEmptyData() *server.AWSConfig {
	return &server.AWSConfig{
		Profiles: []server.AWSProfile{},
	}
}
