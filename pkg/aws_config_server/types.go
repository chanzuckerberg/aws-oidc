package aws_config_server

import (
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

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

func (a *AWSConfig) GetAccounts() []AWSAccount {
	sets.StringSet{}
}

type AWSProfile struct {
	ClientID   okta.ClientID `json:"client_id,omitempty"`
	AWSAccount AWSAccount    `json:"aws_account,omitempty"`
	RoleARN    string        `json:"role_arn,omitempty"`
	IssuerURL  string        `json:"issuer_url,omitempty"`
}

type AWSAccount struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
