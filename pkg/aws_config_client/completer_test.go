package aws_config_client

import (
	"bytes"
	"fmt"
	"testing"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

type testWriter struct {
	bytes.Buffer
}

func (t *testWriter) MergeAWSConfigs(new *ini.File, old *ini.File) (*ini.File, error) {
	return mergeAWSConfigs(new, old)
}

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
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

`
	prompt := &MockPrompt{}
	prompt.inputs = append(prompt.inputs,
		"test-region", // Please input your default AWS region
		1,             // How would you like to configure your AWS config? Configure one role at a time (advanced)

		1,     // Select the AWS account you would like to configure for this profile:
		0,     // What role would you like to use with this profile?
		"",    // what would you like to name this profile? (use default value)
		false, // would you like to configure another account?

		true, // Does this config file look right?
	)

	c := NewCompleter(prompt, generateDummyData(), "", "")

	testWriter := &testWriter{}
	err = c.Complete(baseAWSConfig, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestSurveyProfiles(t *testing.T) {
	r := require.New(t)

	// note how: "Account Name With Spaces" => "account-name-with-spaces"
	expected := `[profile test1]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile account-name-with-spaces]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile my-second-new-profile]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

`

	baseAWSConfig := ini.Empty()

	prompt := &MockPrompt{}
	prompt.inputs = append(prompt.inputs,
		"test-region", // Please input your default AWS region
		1,             // How would you like to configure your AWS config? (Configure 1 role at a time)

		1,    // Select the AWS account you would like to configure for this profile:
		0,    // What role would you like to use with this profile?
		"",   // what would you like to name this profile? (use default value)
		true, // would you like to configure another account?

		0,    // Select the AWS account you would like to configure for this profile:
		0,    // What role would you like to use with this profile?
		"",   // what would you like to name this profile? (use default value)
		true, // would you like to configure another account?

		1,                       // Select the AWS account you would like to configure for this profile:
		0,                       // What role would you like to use with this profile?
		"my-second-new-profile", // what would you like to name this profile?
		false,                   // would you like to configure another account?

		true, // Does this config file look right?
	)

	c := NewCompleter(prompt, generateDummyData(), "", "")

	testWriter := &testWriter{}
	err := c.Complete(baseAWSConfig, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestSurveyRoles(t *testing.T) {
	r := require.New(t)

	expected := `[profile account-name-with-spaces-test1RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile account-name-with-spaces]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile account-name-with-spaces-test2RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test2RoleName
region             = test-region

[profile test1-test1RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile test1]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile test1-test2RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test2RoleName
region             = test-region

`
	newAWSProfiles := ini.Empty()

	prompt := &MockPrompt{}
	prompt.inputs = append(prompt.inputs,
		"test-region", // Please input your default AWS region
		0,             // How would you like to configure your AWS config? (Automatically configure the same role for each account)
		0,             // Select the AWS role you would like to make default
		true,          // Does this config file look right?
	)

	c := NewCompleter(prompt, generateDummyData(), "", "")

	testWriter := &testWriter{}
	err := c.Complete(newAWSProfiles, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestNoSurvey(t *testing.T) {
	r := require.New(t)

	expected := `[profile account-name-with-spaces-test1RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile account-name-with-spaces]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile account-name-with-spaces-test2RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=test2RoleName
region             = test-region

[profile test1-test1RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile test1]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test1RoleName
region             = test-region

[profile test1-test2RoleName]
output             = json
credential_process = aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=test2RoleName
region             = test-region

`
	newAWSProfiles := ini.Empty()

	prompt := &MockPrompt{}
	prompt.inputs = append(prompt.inputs,
		true, // Does this config file look right?
	)

	c := NewCompleter(prompt, generateDummyData(), "test-region", "test1RoleName")

	testWriter := &testWriter{}
	err := c.Complete(newAWSProfiles, testWriter)
	r.NoError(err)
	r.Equal(expected, testWriter.String())
}

func TestNoRoles(t *testing.T) {
	r := require.New(t)
	expected := ``
	baseAWSConfig := ini.Empty()
	prompt := &MockPrompt{}

	c := NewCompleter(prompt, generateEmptyData(), "", "")
	testWriter := &testWriter{}
	err := c.Complete(baseAWSConfig, testWriter)
	r.Error(err)
	r.Equal(expected, testWriter.String())
	r.Equal(baseAWSConfig, ini.Empty())

	// go through the same process with data in baseAWSConfig
	dummySection, err := baseAWSConfig.NewSection("profile acct1")
	r.NoError(err)
	dummySection.NewKey("region", "dummy region")
	err = c.Complete(baseAWSConfig, testWriter)
	r.Error(err)
	acct1Section, err := baseAWSConfig.GetSection("profile acct1")
	r.NoError(err)
	r.Equal(acct1Section, dummySection)
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

	c := NewCompleter(nil, generateDummyData(), "", "")
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

	c := NewCompleter(nil, generateDummyData(), "", "")
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
