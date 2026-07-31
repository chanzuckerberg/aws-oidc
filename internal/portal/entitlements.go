package portal

import (
	"context"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

// Entitlements is everything a person may grant an agent: the accounts and roles they can
// already assume. It is the ceiling for any agent they own. Accounts is the shared access
// model; allowed indexes the same data for O(1) validation of a submitted grant.
type Entitlements struct {
	Accounts []awsaccess.Account
	allowed  map[string]agentsv1.AWSGrant
}

// Allows reports whether the person can grant this account plus role, returning the fully
// populated AWS grant when they can.
func (e *Entitlements) Allows(accountID, roleARN string) (agentsv1.AWSGrant, bool) {
	g, ok := e.allowed[accountID+"|"+roleARN]
	return g, ok
}

// Empty reports whether the person has no grantable access at all.
func (e *Entitlements) Empty() bool {
	return len(e.Accounts) == 0
}

// ResolveEntitlements computes a person's grantable access as a projection of the shared
// access model, so the portal and the config server never disagree about what a person can
// reach.
func ResolveEntitlements(ctx context.Context, sub string, apps okta.AppLister, mappings okta.OIDCRoleMappingsByKey) (*Entitlements, error) {
	access, err := awsaccess.Resolve(ctx, sub, apps, mappings)
	if err != nil {
		return nil, err
	}

	allowed := map[string]agentsv1.AWSGrant{}
	for _, acct := range access.Accounts {
		for _, role := range acct.Roles {
			allowed[acct.ID+"|"+role.RoleARN] = agentsv1.AWSGrant{
				AccountID:    acct.ID,
				AccountAlias: acct.Alias,
				RoleARN:      role.RoleARN,
				RoleName:     role.RoleName,
			}
		}
	}

	return &Entitlements{Accounts: access.Accounts, allowed: allowed}, nil
}
