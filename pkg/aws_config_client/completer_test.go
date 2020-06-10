package aws_config_client

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestLoop(t *testing.T) {
	r := require.New(t)

	// note how: "Account Name With Spaces" => "account-name-with-spaces"
	expected := `[profile account-name-with-spaces]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=foo_client_id --aws-role-arn=arn::::foo_id_1:foo1RoleName 2> /dev/tty'

[profile my-second-new-profile]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=arn::::test_id_1:test1RoleName 2> /dev/tty'

[profile test1]
output             = json
credential_process = sh -c 'aws-oidc creds-process --issuer-url=issuer-url --client-id=bar_client_id --aws-role-arn=arn::::test_id_1:test2RoleName 2> /dev/tty'

`

	out := ini.Empty()
	prompt := &MockPrompt{

		selectResponse: []int{
			0, 0, // select the first role in the first account
			1, 0, // select the first role in the second account
			1, 1, // select the second role in the second account
		},
		inputResponse:   []string{"", "my-second-new-profile", ""},
		confirmResponse: []bool{true, true, false},
	}

	c := NewCompleter(prompt, generateDummyData(), "issuer-url")

	err := c.Loop(out)
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

	c := NewCompleter(nil, generateDummyData(), "issuer-url")
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

// For now generate dummy data, will later on use this for tests instead
func generateDummyData() map[okta.ClientID][]server.ConfigProfile {
	configProfile1 := []server.ConfigProfile{
		{
			AcctName: "test1",
			RoleARN: arn.ARN{
				AccountID: "test_id_1",
				Resource:  "test1RoleName",
			},
		},
		{
			AcctName: "test1",
			RoleARN: arn.ARN{
				AccountID: "test_id_1",
				Resource:  "test2RoleName",
			},
		},
	}
	configProfile2 := []server.ConfigProfile{
		{
			AcctName: "Account Name With Spaces",
			RoleARN: arn.ARN{
				AccountID: "foo_id_1",
				Resource:  "foo1RoleName",
			},
		},
		{
			AcctName: "Account Name With Spaces",
			RoleARN: arn.ARN{
				AccountID: "foo_id_1",
				Resource:  "foo2RoleName",
			},
		},
	}

	data := map[okta.ClientID][]server.ConfigProfile{}
	data["bar_client_id"] = configProfile1
	data["foo_client_id"] = configProfile2
	return data
}
