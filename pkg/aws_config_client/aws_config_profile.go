package aws_config_client

import "github.com/chanzuckerberg/aws-oidc/pkg/okta"

type AWSConfigProfile struct {
	Name    string
	RoleARN string

	ClientID okta.ClientID
}
