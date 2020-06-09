package aws_config_client

import server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"

type AWSConfigProfile struct {
	Name    string
	RoleARN string

	ClientID server.ClientID
}
