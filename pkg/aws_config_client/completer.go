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

func (c *completer) getAccountOptions(accounts []server.AWSAccount) []string {
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
		roleOptions = append(roleOptions, profile.RoleName)
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
func (c *completer) SurveyProfile() ([]*AWSNamedProfile, error) {
	// first prompt for account
	accounts := c.awsConfig.GetAccounts()
	accountIdx, err := c.prompt.Select(
		"Select the AWS account you would like to configure for this profile:",
		c.getAccountOptions(accounts),
	)
	if err != nil {
		return nil, err
	}
	account := accounts[accountIdx]

	// now ask for a role in that account
	profiles := c.awsConfig.GetProfilesForAccount(account)
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
		c.calculateDefaultProfileName(account),
		survey.WithValidator(c.awsProfileNameValidator),
	)
	if err != nil {
		return nil, err
	}

	namedProfile := &AWSNamedProfile{
		Name:       profileName,
		AWSProfile: profile,
	}

	return []*AWSNamedProfile{namedProfile}, nil
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
	targetRole := roles[roleIdx]

	configuredProfiles := []*AWSNamedProfile{}

	for _, account := range accounts {
		profileName := c.calculateDefaultProfileName(account)

		// get the roles associated with this account
		profiles := c.awsConfig.GetProfilesForAccount(account)
		for _, profile := range profiles {

			// Initialize a new AWSNamedProfile
			currentProfile := AWSNamedProfile{
				AWSProfile: server.AWSProfile{
					ClientID:   profile.ClientID,
					AWSAccount: profile.AWSAccount,
					RoleARN:    profile.RoleARN,
					IssuerURL:  profile.IssuerURL,
				},
			}

			currentProfile.Name = fmt.Sprintf("%s-%s", profileName, profile.RoleName)
			configuredProfiles = append(configuredProfiles, &currentProfile)

			if profile.RoleName == targetRole {
				defaultProfile := AWSNamedProfile{
					Name: profileName,
					AWSProfile: server.AWSProfile{
						ClientID:   profile.ClientID,
						AWSAccount: profile.AWSAccount,
						RoleARN:    profile.RoleARN,
						IssuerURL:  profile.IssuerURL,
					},
				}
				configuredProfiles = append(configuredProfiles, &defaultProfile)
			}
		}
	}
	return configuredProfiles, nil
}

func (c *completer) Survey() ([]*AWSNamedProfile, error) {
	configureOptions := []string{"Role", "Profile"}
	configureIdx, err := c.prompt.Select("Would you like to configure AWS by role or profile?", configureOptions)
	if err != nil {
		return nil, err
	}
	configureMethod := configureOptions[configureIdx]

	if configureMethod == "Role" {
		return c.SurveyRoles()
	}
	return c.SurveyProfile()
}

func (c *completer) Continue() (bool, error) {
	return c.prompt.Confirm("Would you like to configure another profile?", true)
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

		// First delete sections with this name so old configuration doesn't persist
		out.DeleteSection(profileSection)
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

func (c *completer) Loop(out *ini.File) error {
	if len(c.awsConfig.Profiles) == 0 {
		logrus.Info("You are not authorized for any roles. Please contact your AWS administrator if this is a mistake")
		return nil
	}

	// assume same region for all accounts configured in this run?
	region, err := c.SurveyRegion()
	if err != nil {
		return err
	}

	for {
		profiles, err := c.Survey()
		if err != nil {
			return err
		}

		err = c.writeAWSProfiles(out, region, profiles)
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
