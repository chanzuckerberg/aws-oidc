package aws_config_client

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

type awsAccount struct {
	id   string
	name string
}

type completer struct {
	awsAccounts map[awsAccount]map[server.ConfigProfile]server.ClientID
	issuerURL   string

	prompt Prompt
}

func NewCompleter(
	prompt Prompt,
	clientMapping map[server.ClientID][]server.ConfigProfile,
	issuerURL string,
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
		issuerURL:   issuerURL,
		prompt:      prompt,
	}
}

func (c *completer) getAccounts() []awsAccount {
	accounts := []awsAccount{}

	for account := range c.awsAccounts {
		accounts = append(accounts, account)
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].name < accounts[j].name
	})

	return accounts
}

func (c *completer) getAccountOptions(accounts []awsAccount) []string {
	accountOptions := []string{}
	for _, account := range accounts {
		accountOptions = append(accountOptions, fmt.Sprintf("%s (%s)", account.name, account.id))
	}
	return accountOptions
}

func (c *completer) getRolesForAccount(account awsAccount) []server.ConfigProfile {
	roles := []server.ConfigProfile{}
	roleToClientIDs := c.awsAccounts[account]

	for role := range roleToClientIDs {
		roles = append(roles, role)
	}

	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].RoleARN.String() < roles[j].RoleARN.String()
	})

	return roles
}

func (c *completer) getRoleOptions(roles []server.ConfigProfile) []string {
	roleOptions := []string{}

	for _, role := range roles {
		roleOptions = append(roleOptions, role.RoleARN.String())
	}
	return roleOptions
}

// Validates that the inputted aws profile name is a valid one
func (c *completer) awsProfileNameValidator(input interface{}) error {
	inputString, ok := input.(string)
	if !ok {
		return errors.New("input not a string")
	}
	valid := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	ok = valid.MatchString(inputString)
	if !ok {
		return errors.Errorf("Input (%s) not a valid AWS profile name", inputString)
	}
	return nil
}

func (c *completer) calculateDefaultProfileName(account awsAccount) string {
	invalid := regexp.MustCompile("[^a-zA-Z0-9_-]")
	replaced := invalid.ReplaceAllString(account.name, "-")
	return strings.ToLower(replaced)
}

func (c *completer) getClientID(account awsAccount, role server.ConfigProfile) server.ClientID {
	return c.awsAccounts[account][role]
}

// SurveyProfile will ask a user to configure an aws profile
func (c *completer) SurveyProfile() (*AWSConfigProfile, error) {
	// first prompt for account
	accounts := c.getAccounts()
	accountIdx, err := c.prompt.Select(
		"Select the AWS account you would like to configure for this profile:",
		c.getAccountOptions(accounts),
	)
	if err != nil {
		return nil, err
	}
	account := accounts[accountIdx]

	// now ask for a role in that account
	roles := c.getRolesForAccount(account)
	roleIdx, err := c.prompt.Select(
		"Select the AWS role you would like to configure for this profile:",
		c.getRoleOptions(roles),
	)
	if err != nil {
		return nil, err
	}
	role := roles[roleIdx]

	// now attempt to name the profile
	profileName, err := c.prompt.Input(
		"What would you like to name this profile:",
		c.calculateDefaultProfileName(account),
		survey.WithValidator(c.awsProfileNameValidator),
	)
	if err != nil {
		return nil, err
	}

	profile := &AWSConfigProfile{
		Name:    profileName,
		RoleARN: role.RoleARN.String(),

		ClientID: c.getClientID(account, role),
	}

	return profile, nil
}

func (c *completer) Continue() (bool, error) {
	return c.prompt.Confirm("Would you like to configure another profile?", true)
}

func (c *completer) writeAWSProfile(out *ini.File, profile *AWSConfigProfile) error {
	profileSection := fmt.Sprintf("profile %s", profile.Name)

	credsProcessValue := fmt.Sprintf(
		"sh -c 'aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s 2> /dev/tty'",
		c.issuerURL,
		profile.ClientID,
		profile.RoleARN,
	)

	section, err := out.NewSection(profileSection)
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s section in AWS Config", profileSection)
	}
	section.Key("output").SetValue("json")
	section.Key("credential_process").SetValue(credsProcessValue)
	return nil
}

func (c *completer) Loop(out *ini.File) error {
	for {
		profile, err := c.SurveyProfile()
		if err != nil {
			return err
		}

		err = c.writeAWSProfile(out, profile)
		if err != nil {
			return err
		}

		cnt, err := c.Continue()
		if err != nil {
			return err
		}
		if !cnt {
			break
		}
	}
	return nil
}
