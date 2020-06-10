package aws_config_server

import "github.com/chanzuckerberg/aws-oidc/pkg/okta"

type AWSConfig struct {
	Profiles []AWSProfile `json:"profiles,omitempty"`
}

func (a *AWSConfig) HasAccount(acctName string) bool {
	for _, profile := range a.Profiles {
		if profile.AWSAccountName == acctName {
			return true
		}
	}
	return false
}

type AWSProfile struct {
	ClientID       okta.ClientID `json:"client_id,omitempty"`
	AWSAccountName string        `json:"aws_account_name,omitempty"`
	RoleARN        string        `json:"role_arn,omitempty"`
	IssuerURL      string        `json:"issuer_url,omitempty"`
}
