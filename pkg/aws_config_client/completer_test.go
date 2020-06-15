package aws_config_client

import (
	"fmt"
	"testing"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/stretchr/testify/require"
)

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

func generateDummyData() *server.AWSConfig {
	return &server.AWSConfig{
		Profiles: []server.AWSProfile{
			{
				ClientID: "bar_client_id",
				AWSAccount: server.AWSAccount{
					Name: "test1",
					ID:   "test_id_1",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
			},
			{
				ClientID: "bar_client_id",
				AWSAccount: server.AWSAccount{
					Name: "test1",
					ID:   "test_id_1",
				},
				RoleARN:   "test2RoleName",
				IssuerURL: "issuer-url",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name: "Account Name With Spaces",
					ID:   "account id 2",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
			},
			{
				ClientID: "foo_client_id",
				AWSAccount: server.AWSAccount{
					Name: "Account Name With Spaces",
					ID:   "account id 2",
				},
				RoleARN:   "test1RoleName",
				IssuerURL: "issuer-url",
			},
		},
	}
}
