package aws_config_client

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var (
	homeDir              = "~"
	DefaultAWSConfigPath = filepath.Join(homeDir, ".aws", "config")
)

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultAWSConfigPath = filepath.Join(homeDir, ".aws", "config")
}

const (
	clientIDRegex     = "--client-id=(?P<ClientID>\\S+)"
	issuerURLRegex    = "--issuer-url=(?P<IssuerURL>\\S+)"
	roleARNRegex      = "--aws-role-arn=(?P<RoleARN>\\S+)"
	FlagProfile       = "profile"
	defaultAWSProfile = "default"
	envAWSProfile     = "AWS_PROFILE"
)

type AWSOIDCConfiguration struct {
	ClientID  string
	IssuerURL string
	RoleARN   string
	Region    *string
	Output    *string
}

// HACK(el): This is not the best, but decided to do this to:
//   - Not add extraneous fields to user's AWS config files
//   - Not have to maintain a parallel config file
func extractCredProcess(command string, regex string) (string, error) {
	r := regexp.MustCompile(regex)
	submatches := r.FindStringSubmatch(command)
	if len(submatches) != 2 {
		return "", fmt.Errorf("did not find match")
	}
	return submatches[1], nil
}

func cleanCredProcessCommand(command string) string {
	// clean up until the first quote
	before := regexp.MustCompile("^.*?['\"]")
	// clean up after the last quote
	after := regexp.MustCompile("['\"].*$")
	// get rid of the tty hack if present
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
			return "", fmt.Errorf("could not read command line flag %s: %w", FlagProfile, err)
		}
		profile = flagProfileValue
	}

	return profile, nil
}

func FetchParamsFromAWSConfig(cmd *cobra.Command, awsConfigPath string) (*AWSOIDCConfiguration, error) {
	ini, err := ini.Load(awsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not open aws config: %w", err)
	}

	// grab the section corresponding to our profile
	profileName, err := resolveProfile(cmd)
	if err != nil {
		return nil, err
	}
	profileSectionName := fmt.Sprintf("profile %s", profileName)
	section, err := ini.GetSection(profileSectionName)
	if err != nil {
		return nil, fmt.Errorf("could not fetch %s section from aws config: %w", profileSectionName, err)
	}

	credProcess, err := section.GetKey(AWSConfigSectionCredentialProcess)
	if err != nil {
		return nil, fmt.Errorf("%s not defined for profile %s: %w", AWSConfigSectionCredentialProcess, profileSectionName, err)
	}

	credProcessCmd := credProcess.String()
	credProcessCmd = cleanCredProcessCommand(credProcessCmd)
	clientID, err := extractCredProcess(credProcessCmd, clientIDRegex)
	if err != nil {
		return nil, fmt.Errorf("could not extract --client-id from (%s): %w", credProcessCmd, err)
	}
	issuerURL, err := extractCredProcess(credProcessCmd, issuerURLRegex)
	if err != nil {
		return nil, fmt.Errorf("could not extract --issuer-url from (%s): %w", credProcessCmd, err)
	}
	roleARN, err := extractCredProcess(credProcessCmd, roleARNRegex)
	if err != nil {
		return nil, fmt.Errorf("could not extract --aws-role-arn from (%s): %w", credProcessCmd, err)
	}

	currentConfig := &AWSOIDCConfiguration{
		ClientID:  clientID,
		IssuerURL: issuerURL,
		RoleARN:   roleARN,
	}
	region, err := section.GetKey("region")
	if err != nil {
		slog.Debug("getting region from aws config", "error", err)
	} else {
		currentConfig.Region = aws.String(region.String())
	}

	output, err := section.GetKey("output")
	if err != nil {
		slog.Debug("getting output from aws config", "error", err)
	} else {
		currentConfig.Output = aws.String(output.String())
	}

	return currentConfig, nil
}
