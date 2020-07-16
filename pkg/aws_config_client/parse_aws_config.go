package aws_config_client

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

const (
	clientIDRegex  = "--client-id=(?P<ClientID>\\S+)"
	issuerURLRegex = "--issuer-url=(?P<IssuerURL>\\S+)"
	roleARNRegex   = "--aws-role-arn=(?P<RoleARN>\\S+)"

	FlagProfile = "profile"

	defaultAWSProfile    = "default"
	DefaultAWSConfigPath = "~/.aws/config"

	envAWSProfile = "AWS_PROFILE"
)

type AWSOIDCConfiguration struct {
	ClientID  string
	IssuerURL string
	RoleARN   string
}

// HACK(el): This is not the best, but decided to do this to:
//           - Not add extraneous fields to user's AWS config files
//           - Not have to maintain a parallel config file
func extractCredProcess(command string, regex string) (string, error) {
	r := regexp.MustCompile(regex)
	submatches := r.FindStringSubmatch(command)
	if len(submatches) != 2 {
		return "", errors.Errorf("did not find match")
	}
	return submatches[1], nil
}

func cleanCredProcessCommand(command string) string {
	// clean up until the first quote
	before := regexp.MustCompile("^.*?['\"]")
	// clean up after the last quote
	after := regexp.MustCompile("['\"].*$")
	// get rid of the tty if present
	tty := regexp.MustCompile("2> /dev/tty")

	command = string(before.ReplaceAll([]byte(command), []byte("")))
	command = string(after.ReplaceAll([]byte(command), []byte("")))
	command = string(tty.ReplaceAll([]byte(command), []byte("")))
	command = strings.TrimSpace(command)

	return command
}

func resolveProfile(cmd *cobra.Command) (string, error) {
	// the default profile
	profile := defaultAWSProfile

	// env has precedence over default
	envProfile, present := os.LookupEnv(envAWSProfile)
	if present {
		profile = envProfile
	}

	// flag has precedence over env and default
	if cmd != nil && cmd.Flags().Changed(FlagProfile) {
		flagProfileValue, err := cmd.Flags().GetString(FlagProfile)
		if err != nil {
			return "", errors.Wrapf(err, "could not read command line flag %s", FlagProfile)
		}
		profile = flagProfileValue
	}

	return profile, nil
}

func FetchParamsFromAWSConfig(cmd *cobra.Command, awsConfigPath string) (*AWSOIDCConfiguration, error) {
	// load and parse AWS config file
	awsConfigPath, err := homedir.Expand(awsConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse aws config file path")
	}
	ini, err := ini.Load(awsConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open aws config")
	}

	// grab the section corresponding to our profile
	profileName, err := resolveProfile(cmd)
	if err != nil {
		return nil, err
	}
	profileSectionName := fmt.Sprintf("profile %s", profileName)
	section, err := ini.GetSection(profileSectionName)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"could not fetch %s section from aws config",
			profileSectionName)
	}

	credProcess, err := section.GetKey(AWSConfigSectionCredentialProcess)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"%s not defined for profile %s",
			AWSConfigSectionCredentialProcess,
			profileSectionName,
		)
	}

	credProcessCmd := credProcess.String()
	credProcessCmd = cleanCredProcessCommand(credProcessCmd)
	clientID, err := extractCredProcess(credProcessCmd, clientIDRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "could not extract --client-id from (%s)", credProcessCmd)
	}
	issuerURL, err := extractCredProcess(credProcessCmd, issuerURLRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "could not extract --issuer-url from (%s)", credProcessCmd)
	}
	roleARN, err := extractCredProcess(credProcessCmd, roleARNRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "could not extract --aws-role-arn from (%s)", credProcessCmd)
	}
	return &AWSOIDCConfiguration{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		RoleARN:   roleARN,
	}, nil
}
