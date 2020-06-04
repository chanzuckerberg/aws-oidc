package aws_config_server

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAWSConfig(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	configFile, err := createAWSConfig(ctx, &testAWSConfigGenerationParams, testConfigMapping, []string{"clientID1", "clientID2", "clientID3"})
	r.NoError(err)
	r.NotEmpty(configFile)

	acct1Header := "profile account1"
	acct1Section, err := configFile.GetSection(acct1Header)
	r.NoError(err)
	r.NotEmpty(acct1Section)

	acct2Header := "profile account2"
	acct2Section, err := configFile.GetSection(acct2Header)
	r.NoError(err)
	r.NotEmpty(acct2Section)

	acct3Header := "profile account3"
	acct3Section, err := configFile.GetSection(acct3Header)
	r.NoError(err)
	r.NotEmpty(acct3Section)

	acct_with_space_header := "profile account-with-space"
	acct_with_space_section, err := configFile.GetSection(acct_with_space_header)
	r.NoError(err)
	r.NotEmpty(acct_with_space_section)

	credsProcessFormat := "sh -c 'aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s 2> /dev/tty'"
	output := "json"

	acct1credsProcess := fmt.Sprintf(credsProcessFormat, testAWSConfigGenerationParams.OIDCProvider, "clientID1", "arn:aws:iam::AccountNumber1:role/WorkerRole")
	r.Equal(acct1Section.Key("output").Value(), output)
	r.Equal(acct1Section.Key("credential_process").Value(), acct1credsProcess)

	acct2credsProcess := fmt.Sprintf(credsProcessFormat, testAWSConfigGenerationParams.OIDCProvider, "clientID1", "arn:aws:iam::AccountNumber2:role/WorkerRole")
	r.Equal(acct2Section.Key("output").Value(), output)
	r.Equal(acct2Section.Key("credential_process").Value(), acct2credsProcess)

	acct3credsProcess := fmt.Sprintf(credsProcessFormat, testAWSConfigGenerationParams.OIDCProvider, "clientID2", "arn:aws:iam::AccountNumber3:role/WorkerRole")
	r.Equal(acct3Section.Key("output").Value(), output)
	r.Equal(acct3Section.Key("credential_process").Value(), acct3credsProcess)
}
