package aws_config_client

import server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"

type AWSNamedProfile struct {
	Name       string
	AWSProfile server.AWSProfile
}
