package aws_config_client

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

const DefaultAWSRegion = "us-west-2"

type completer struct {
	awsConfig *server.AWSConfig
	prompt    Prompt
}

func NewCompleter(
	prompt Prompt,
	awsConfig *server.AWSConfig,
) *completer {

	return &completer{
		awsConfig: awsConfig,
		prompt:    prompt,
	}
}

func (c *completer) getAccountOptions(accounts []*server.AWSAccount) []string {
	accountOptions := []string{}
	for _, account := range accounts {
		accountOptions = append(
			accountOptions,
			fmt.Sprintf("%s (%s)", account.Name, account.ID))
	}
	return accountOptions
}

func (c *completer) getRoleOptions(profiles []server.AWSProfile) []string {
	roleOptions := []string{}
	for _, profile := range profiles {
		roleOptions = append(roleOptions, profile.RoleARN)
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

func (c *completer) calculateDefaultProfileName(account server.AWSAccount) string {
	invalid := regexp.MustCompile("[^a-zA-Z0-9_-]")
	replaced := invalid.ReplaceAllString(account.Name, "-")
	return strings.ToLower(replaced)
}

func (c *completer) SurveyRegion() (string, error) {
	return c.prompt.Input(
		"Please input your default AWS region:",
		DefaultAWSRegion,
	)
}

// SurveyProfile will ask a user to configure an aws profile
func (c *completer) SurveyRemainingProfiles(consumedProfiles []*AWSNamedProfile, remainingAccounts []*server.AWSAccount) ([]*server.AWSAccount, error) {
	// first prompt for account
	// accounts := c.awsConfig.GetAccounts()
	accountIdx, err := c.prompt.Select(
		"Select the AWS account you would like to configure for this profile:",
		c.getAccountOptions(remainingAccounts),
	)
	if err != nil {
		return nil, err
	}
	account := remainingAccounts[accountIdx]

	// now ask for a role in that account
	profiles := c.awsConfig.GetProfilesForAccount(*account)
	profileIdx, err := c.prompt.Select(
		"Select the AWS role you would like to configure for this profile:",
		c.getRoleOptions(profiles),
	)
	if err != nil {
		return nil, err
	}
	profile := profiles[profileIdx]

	// now attempt to name the profile
	profileName, err := c.prompt.Input(
		"What would you like to name this profile:",
		c.calculateDefaultProfileName(*account),
		survey.WithValidator(c.awsProfileNameValidator),
	)
	if err != nil {
		return nil, err
	}

	namedProfile := &AWSNamedProfile{
		Name:       profileName,
		AWSProfile: profile,
	}
	consumedProfiles = append(consumedProfiles, namedProfile)

	// delete account from remainingAccounts
	remainingAccounts[accountIdx] = remainingAccounts[len(remainingAccounts)-1]
	remainingAccounts = remainingAccounts[:len(remainingAccounts)-1]

	return remainingAccounts, nil
}

func (c *completer) consumeProfiles(targetRoleName string, accounts []server.AWSAccount) ([]*AWSNamedProfile, []*server.AWSAccount) {
	consumedProfiles := []*AWSNamedProfile{}
	remainingAccounts := []*server.AWSAccount{}

	for _, account := range accounts {
		profileName := c.calculateDefaultProfileName(account)

		// get the roles associated with this account
		selectedProfile := server.AWSProfile{}
		profiles := c.awsConfig.GetProfilesForAccount(account)
		for _, profile := range profiles {
			if strings.Contains(profile.RoleARN, targetRoleName) {
				selectedProfile = profile
			}
		}
		// if roleARN is populated, then we add it to the consumedProfiles list
		if selectedProfile.RoleARN == "" {
			currentProfile := &AWSNamedProfile{
				Name:       profileName,
				AWSProfile: selectedProfile,
			}
			consumedProfiles = append(consumedProfiles, currentProfile)
			continue
		}
		// Else, then populate the remainingAccounts list
		remainingAccounts = append(remainingAccounts, &account)
	}

	return consumedProfiles, remainingAccounts
}

// SurveyRole will ask a user to configure a default role
func (c *completer) SurveyRoles() ([]*AWSNamedProfile, error) {
	// first prompt for roles
	roles := c.awsConfig.GetRoleNames()
	accounts := c.awsConfig.GetAccounts()

	roleIdx, err := c.prompt.Select(
		"Select the AWS role you would like to make default:",
		roles,
	)
	if err != nil {
		return nil, err
	}
	role := roles[roleIdx]

	// Next: configure all the roles for all the profiles
	configuredProfiles, remainingAccounts := c.consumeProfiles(role, accounts)

	for {
		// Ask user if they want to configure more profiles
		cnt, err := c.ContinueRemaining(remainingAccounts)
		if err != nil {
			return nil, err
		}
		if !cnt {
			// Will exit if len(remainingProfiles) == 0
			return configuredProfiles, nil
		}
		remainingAccounts, err = c.SurveyRemainingProfiles(configuredProfiles, remainingAccounts)
		if err != nil {
			return configuredProfiles, err
		}
	}
}

func (c *completer) Continue() (bool, error) {
	return c.prompt.Confirm("Would you like to configure another profile?", true)
}

func (c *completer) ContinueRemaining(remainingAccounts []*server.AWSAccount) (bool, error) {
	if len(remainingAccounts) == 0 {
		return false, nil
	}
	acctNames := []string{}
	for _, acct := range remainingAccounts {
		acctNames = append(acctNames, acct.Name)
	}

	confirmPrompt := fmt.Sprintf("Would you like to configure the rest?\nRemaining Accounts: %v", acctNames)
	return c.prompt.Confirm(confirmPrompt, true)
}

func (c *completer) writeAWSProfile(out *ini.File, region string, profile *AWSNamedProfile) error {
	profileSection := fmt.Sprintf("profile %s", profile.Name)

	credsProcessValue := fmt.Sprintf(
		"sh -c 'aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s 2> /dev/tty'",
		profile.AWSProfile.IssuerURL,
		profile.AWSProfile.ClientID,
		profile.AWSProfile.RoleARN,
	)

	section, err := out.NewSection(profileSection)
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s section in AWS Config", profileSection)
	}
	section.Key("output").SetValue("json")
	section.Key("credential_process").SetValue(credsProcessValue)
	section.Key("region").SetValue(region)
	return nil
}

func (c *completer) writeAWSProfiles(out *ini.File, region string, profiles []*AWSNamedProfile) error {
	for _, profile := range profiles {
		profileSection := fmt.Sprintf("profile %s", profile.Name)

		credsProcessValue := fmt.Sprintf(
			"sh -c 'aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s 2> /dev/tty'",
			profile.AWSProfile.IssuerURL,
			profile.AWSProfile.ClientID,
			profile.AWSProfile.RoleARN,
		)

		section, err := out.NewSection(profileSection)
		if err != nil {
			return errors.Wrapf(err, "Unable to create %s section in AWS Config", profileSection)
		}
		section.Key("output").SetValue("json")
		section.Key("credential_process").SetValue(credsProcessValue)
		section.Key("region").SetValue(region)
	}

	return nil
}

// Name is misleading, but kept it for compile-bility
func (c *completer) Loop(out *ini.File) error {
	// assume same region for all accounts configured in this run?
	region, err := c.SurveyRegion()
	if err != nil {
		return err
	}

	// This is where we loop indefinitely until we're done
	profiles, err := c.SurveyRoles()
	if err != nil {
		return err
	}

	err = c.writeAWSProfiles(out, region, profiles)
	if err != nil {
		return err
	}

	return nil
}
