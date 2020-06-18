package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var profileName string
var sessionDuration time.Duration

const (
	clientIDRegex  = "--client-id=(?P<ClientID>\\S+)"
	issuerURLRegex = "--issuer-url=(?P<IssuerURL>\\S+)"
	roleARNRegex   = "--aws-role-arn=(?P<RoleARN>\\S+)"
)

func init() {
	envCmd.Flags().StringVar(&profileName, "profile", "", "AWS Profile to fetch credentials from.")
	envCmd.MarkFlagRequired("profile") //nolint:errcheck

	envCmd.Flags().DurationVar(
		&sessionDuration,
		"session-duration",
		time.Hour,
		"The duration, of the role session. Must be between 1-12 hours. `1h` means 1 hour."
	)

	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "aws-oidc env",
	Long: `Env will output relevant AWS credentials to environment variables.
	Useful when running docker such "docker run -it --env-file <(aws-oidc env --profile foobar) amazon/aws-cli sts get-caller-identity"
	`,
	RunE: envRun,
}

func envRun(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// load and parse AWS config file
	awsConfigPath, err := homedir.Expand("~/.aws/config")
	if err != nil {
		return errors.Wrap(err, "Could not parse aws config file path")
	}
	ini, err := ini.Load(awsConfigPath)
	if err != nil {
		return errors.Wrap(err, "could not open aws config")
	}

	profileSectionName := fmt.Sprintf("profile %s", profileName)
	section, err := ini.GetSection(profileSectionName)
	if err != nil {
		return errors.Wrapf(
			err,
			"could not fetch %s from aws config",
			profileSectionName)
	}

	credProcess, err := section.GetKey(aws_config_client.AWSConfigSectionCredentialProcess)
	if err != nil {
		return errors.Wrapf(
			err,
			"%s not defined for profile %s",
			aws_config_client.AWSConfigSectionCredentialProcess,
			profileSectionName,
		)
	}

	credProcessCmd := credProcess.String()

	clientID, err = extractCredProcess(credProcessCmd, clientIDRegex)
	if err != nil {
		return errors.Wrapf(err, "could not extract --client-id from (%s)", credProcessCmd)
	}
	issuerURL, err = extractCredProcess(credProcessCmd, issuerURLRegex)
	if err != nil {
		return errors.Wrapf(err, "could not extract --issuer-url from (%s)", credProcessCmd)
	}
	roleARN, err = extractCredProcess(credProcessCmd, roleARNRegex)
	if err != nil {
		return errors.Wrapf(err, "could not extract --aws-role-arn from (%s)", credProcessCmd)
	}

	assumeRoleOutput, err := assumeRole(
		ctx,
		clientID,
		issuerURL,
		roleARN,
	)
	if err != nil {
		return nil
	}

	envVars := getAWSEnvVars(assumeRoleOutput)

	// output in the appropriate format for docker
	fmt.Fprintln(os.Stdout, strings.Join(envVars, "\n"))
	return nil
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
