package aws

import (
	"context"
	"fmt"
	"sync"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	defaultProvisionerRoleName = "agent-provisioner"
	// defaultSessionDuration is the AWS minimum for AssumeRole. Cross-account sessions are
	// kept this short on purpose: the operator only needs them briefly to create roles.
	defaultSessionDuration = 15 * time.Minute
	defaultRegion          = "us-east-1"
	sessionName            = "aws-oidc-agent-operator"
)

// AssumeRoleConfig configures cross-account IAM client creation.
type AssumeRoleConfig struct {
	// ProvisionerRoleName is the role assumed in each target account (the registry knows
	// these roles exist and trusts the operator's control-plane role to assume them).
	ProvisionerRoleName string
	// SessionDuration bounds each cross-account session. Keep it short.
	SessionDuration time.Duration
	// Region is the region STS and IAM calls use. IAM is global, but the SDK requires a
	// region to resolve an endpoint.
	Region string
}

// assumeRoleFactory hands out IAM clients scoped to a target account by assuming that
// account's provisioner role from the operator's own IRSA identity. It caches one client per
// account; the SDK's credential cache refreshes the short-lived session under each client, so
// concurrent grants in the same account share a single assume-role session. It is safe for
// concurrent use.
type assumeRoleFactory struct {
	stsClient           *sts.Client
	baseCfg             awssdk.Config
	provisionerRoleName string
	sessionDuration     time.Duration

	mu      sync.Mutex
	clients map[string]IAMAPI
}

// NewAssumeRoleClientFactory builds a ClientFactory backed by cross-account role assumption.
// The operator's ambient credentials (its IRSA role, attached by annotation) are the base
// identity; that role is trusted to assume each account's provisioner role.
func NewAssumeRoleClientFactory(ctx context.Context, cfg AssumeRoleConfig) (ClientFactory, error) {
	if cfg.ProvisionerRoleName == "" {
		cfg.ProvisionerRoleName = defaultProvisionerRoleName
	}
	if cfg.SessionDuration == 0 {
		cfg.SessionDuration = defaultSessionDuration
	}
	if cfg.Region == "" {
		cfg.Region = defaultRegion
	}

	baseCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("loading operator aws config: %w", err)
	}

	f := &assumeRoleFactory{
		stsClient:           sts.NewFromConfig(baseCfg),
		baseCfg:             baseCfg,
		provisionerRoleName: cfg.ProvisionerRoleName,
		sessionDuration:     cfg.SessionDuration,
		clients:             map[string]IAMAPI{},
	}
	return f.iamFor, nil
}

// iamFor returns an IAM client scoped to the target account, building and caching one on
// first use. It satisfies ClientFactory.
func (f *assumeRoleFactory) iamFor(_ context.Context, accountID string) (IAMAPI, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	client, ok := f.clients[accountID]
	if ok {
		return client, nil
	}

	roleARN := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, f.provisionerRoleName)
	provider := stscreds.NewAssumeRoleProvider(f.stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
		o.Duration = f.sessionDuration
		o.RoleSessionName = sessionName
	})

	accountCfg := f.baseCfg.Copy()
	accountCfg.Credentials = awssdk.NewCredentialsCache(provider)

	client = iam.NewFromConfig(accountCfg)
	f.clients[accountID] = client
	return client, nil
}
