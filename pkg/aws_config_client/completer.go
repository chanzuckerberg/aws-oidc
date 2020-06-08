package aws_config_client

import (
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
)

type awsAccount struct {
	id   string
	name string
}

type completer struct {
	awsAccounts map[awsAccount]map[server.ConfigProfile]server.ClientID
}

func NewCompleter(
	clientMapping map[server.ClientID][]server.ConfigProfile,
) *completer {

	awsAccounts := map[awsAccount]map[server.ConfigProfile]server.ClientID{}

	// Invert the map to help generate configs
	for clientID, configProfiles := range clientMapping {
		for _, configProfile := range configProfiles {
			awsAccount := awsAccount{
				name: configProfile.AcctName,
				id:   configProfile.RoleARN.AccountID,
			}

			awsRoles, ok := awsAccounts[awsAccount]
			if !ok {
				awsRoles = map[server.ConfigProfile]server.ClientID{}
			}
			awsRoles[configProfile] = clientID
			awsAccounts[awsAccount] = awsRoles
		}
	}

	return &completer{
		awsAccounts: awsAccounts,
	}
}

func (c *completer) getAccounts() []awsAccount {
	accounts := []awsAccount{}

	for account := range c.awsAccounts {
		accounts = append(accounts, account)
	}
	return accounts
}

func (c *completer) getRolesForAccount(account awsAccount) []server.ConfigProfile {
	roles := []server.ConfigProfile{}
	for configProfile := range c.awsAccounts[account] {
		roles = append(roles, configProfile)
	}
	return roles
}

func (c *completer) getClientID(account awsAccount, profile server.ConfigProfile) server.ClientID {
	return c.awsAccounts[account][profile]
}

func (c *completer) getAccountsSuggestions() []string {
	accounts := c.getAccounts()

	suggests := []string{}

	for _, account := range accounts {
		suggests = append(suggests, account.name)
	}
	return suggests
}

func (c *completer) CompleteAccount() []string {
	return c.getAccountsSuggestions()
}
