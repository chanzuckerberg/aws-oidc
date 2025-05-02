package aws_config_client

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

const DefaultAWSRegion = "us-west-2"

type completer struct {
	awsConfig       *server.AWSConfig
	prompt          Prompt
	defaultRegion   string
	defaultRoleName string
}

func NewCompleter(
	prompt Prompt,
	awsConfig *server.AWSConfig,
	defaultRegion string,
	defaultRoleName string,
) *completer {

	return &completer{
		awsConfig:       awsConfig,
		prompt:          prompt,
		defaultRegion:   defaultRegion,
		defaultRoleName: defaultRoleName,
	}
}

func (c *completer) getAccountOptions(accounts []server.AWSAccount) []string {
	accountOptions := []string{}
	for _, account := range accounts {
		accountOptions = append(
			accountOptions,
			fmt.Sprintf("%s (%s)", account.GetAliasOrName(), account.ID))
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
	replaced := invalid.ReplaceAllString(account.GetAliasOrName(), "-")
	return strings.ToLower(replaced)
}

func (c *completer) SurveyRegion() (string, error) {
	if c.defaultRegion != "" {
		return c.defaultRegion, nil
	}

	return c.prompt.Input(
		"Please input your default AWS region:",
		DefaultAWSRegion,
	)
}

// SurveyProfile will ask a user to configure an aws profile
func (c *completer) SurveyProfile() (*AWSNamedProfile, error) {
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

	return namedProfile, nil
}

// SurveyRole will ask a user to configure a default role
func (c *completer) SurveyRoles() ([]*AWSNamedProfile, error) {
	// first prompt for roles
	roles := c.awsConfig.GetRoleNames()
	accounts := c.awsConfig.GetAccounts()

	var targetRole string

	if c.defaultRoleName == "" {
		roleIdx, err := c.prompt.Select(
			"Select the AWS role you would like to make default:",
			roles,
		)
		if err != nil {
			return nil, err
		}
		targetRole = roles[roleIdx]
	} else {
		targetRole = c.defaultRoleName
	}

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

func (c *completer) SurveyProfiles() ([]*AWSNamedProfile, error) {
	collectedProfiles := []*AWSNamedProfile{}
	for {
		currentProfile, err := c.SurveyProfile()
		if err == terminal.InterruptErr {
			slog.Info("Process Interrupted.")
			break
		}
		if err != nil {
			return nil, err
		}
		collectedProfiles = append(collectedProfiles, currentProfile)
		cnt, err := c.Continue()
		if err != nil {
			return nil, err
		}
		if !cnt {
			break
		}
	}
	return collectedProfiles, nil
}

func (c *completer) Survey() ([]*AWSNamedProfile, error) {
	if c.defaultRoleName != "" {
		return c.SurveyRoles()

	}
	configureOptions := []string{
		"Automatically configure the same role for each account? (recommended)",
		"Configure one role at a time? (advanced)"}
	configureFuncs := []func() ([]*AWSNamedProfile, error){c.SurveyRoles, c.SurveyProfiles}
	configureIdx, err := c.prompt.Select("How would you like to configure your AWS config?", configureOptions)
	if err != nil {
		return nil, err
	}
	return configureFuncs[configureIdx]()
}

func (c *completer) Continue() (bool, error) {
	return c.prompt.Confirm("Would you like to configure another profile?", true)
}

func (c *completer) assembleAWSConfig(region string, profiles []*AWSNamedProfile) (*ini.File, error) {
	out := ini.Empty()

	for _, profile := range profiles {
		profileSection := fmt.Sprintf("profile %s", profile.Name)

		credsProcessValue := fmt.Sprintf(
			"aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s",
			profile.AWSProfile.IssuerURL,
			profile.AWSProfile.ClientID,
			profile.AWSProfile.RoleARN,
		)

		// First delete sections with this name so old configuration doesn't persist
		out.DeleteSection(profileSection)
		section, err := out.NewSection(profileSection)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to create %s section in AWS Config", profileSection)
		}
		section.Key(AWSConfigSectionOutput).SetValue("json")
		section.Key(AWSConfigSectionCredentialProcess).SetValue(credsProcessValue)
		section.Key(AWSConfigSectionRegion).SetValue(region)
	}
	return out, nil
}

func (c *completer) Complete(base *ini.File, awsConfigWriter AWSConfigWriter) error {
	if len(c.awsConfig.Profiles) == 0 {
		return errors.Errorf("You are not authorized for any AWS roles. Please contact your AWS administrator if this is a mistake")
	}

	// ask for a region, assume all profiles configured with this region
	region, err := c.SurveyRegion()
	if err != nil {
		return err
	}

	// figure out the profiles
	profiles, err := c.Survey()
	if err != nil {
		return err
	}

	newAWSProfiles, err := c.assembleAWSConfig(region, profiles)
	if err != nil {
		return err
	}

	mergedConfig, err := awsConfigWriter.MergeAWSConfigs(newAWSProfiles, base)
	if err != nil {
		return err
	}

	_, err = mergedConfig.WriteTo(awsConfigWriter)
	return fmt.Errorf("Could not write new aws config: %w", err)
}
