package aws_config_server

import (
	"context"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
)

func TestCreateAWSConfig(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	config, err := createAWSConfig(
		ctx,
		&testAWSConfigGenerationParams,
		testConfigMapping,
		[]okta.ClientID{"clientID1", "clientID2", "clientID3"},
	)
	r.NoError(err)
	r.NotEmpty(config)

	for _, accountName := range []string{"Account1", "Account2", "Account3"} {
		r.True(config.HasAccount(accountName))

	}
}
