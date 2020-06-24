package aws_config_server

import (
	"sort"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

type AWSConfig struct {
	Profiles []AWSProfile `json:"profiles,omitempty"`
}

func (a *AWSConfig) HasAccount(acctName string) bool {
	for _, profile := range a.Profiles {
		if profile.AWSAccount.Name == acctName {
			return true
		}
	}
	return false
}

func (a *AWSConfig) GetAccounts() []AWSAccount {
	set := map[AWSAccount]struct{}{}
	for _, profile := range a.Profiles {
		set[profile.AWSAccount] = struct{}{}
	}

	accounts := []AWSAccount{}
	for account := range set {
		accounts = append(accounts, account)
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].Name < accounts[j].Name
	})
	return accounts
}

func (a *AWSConfig) GetRoleNames() []string {
	set := map[string]bool{}
	roleNames := []string{}

	for _, profile := range a.Profiles {
		roleName := profile.RoleName
		if _, ok := set[roleName]; !ok {
			set[roleName] = true
			roleNames = append(roleNames, roleName)
		}
	}
	return roleNames
}

func (a *AWSConfig) GetProfilesForAccount(account AWSAccount) []AWSProfile {
	profiles := []AWSProfile{}

	for _, profile := range a.Profiles {
		if profile.AWSAccount == account {
			profiles = append(profiles, profile)
		}
	}

	sort.SliceStable(profiles, func(i, j int) bool {
		return profiles[i].RoleARN < profiles[j].RoleARN
	})

	return profiles
}

type AWSProfile struct {
	ClientID   okta.ClientID `json:"client_id,omitempty"`
	AWSAccount AWSAccount    `json:"aws_account,omitempty"`
	RoleARN    string        `json:"role_arn,omitempty"`
	IssuerURL  string        `json:"issuer_url,omitempty"`
	RoleName   string        `json:"role_name,omitempty"`
}

type AWSAccount struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Alias string `json:"alias,omitempty"`
}
