package portal

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

// MappingsProvider returns the current rolemap grouped by client ID. It is called per
// request so entitlements always reflect the latest rolemap.
type MappingsProvider func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error)

// RoleAccess is a single role the user can assume in an account.
type RoleAccess struct {
	RoleARN  string
	RoleName string
}

// AccountAccess is the set of roles the user can assume in one account.
type AccountAccess struct {
	AccountID    string
	AccountAlias string
	Roles        []RoleAccess
}

// Label is a display name for the account.
func (a AccountAccess) Label() string {
	if a.AccountAlias != "" {
		return a.AccountAlias
	}
	return a.AccountID
}

// Entitlements is everything a user is allowed to grant an agent: the accounts and roles
// they can already assume. It is the ceiling for any agent the user owns.
type Entitlements struct {
	Accounts []AccountAccess
	allowed  map[string]Grant
}

// Allows reports whether the user can grant this account plus role, returning the fully
// populated grant when they can.
func (e *Entitlements) Allows(accountID, roleARN string) (Grant, bool) {
	g, ok := e.allowed[accountID+"|"+roleARN]
	return g, ok
}

// Empty reports whether the user has no grantable access at all.
func (e *Entitlements) Empty() bool {
	return len(e.Accounts) == 0
}

// ResolveEntitlements computes a user's grantable access. It lists the Okta apps assigned
// to the user (their client IDs) and looks each up in the rolemap, exactly as the config
// server does when it builds a human's aws config.
func ResolveEntitlements(ctx context.Context, sub string, apps okta.AppLister, mappings okta.OIDCRoleMappingsByKey) (*Entitlements, error) {
	clientIDs, err := okta.GetClientIDs(ctx, sub, apps)
	if err != nil {
		return nil, fmt.Errorf("getting client IDs for %s: %w", sub, err)
	}

	byAccount := map[string]*AccountAccess{}
	allowed := map[string]Grant{}

	for _, clientID := range clientIDs {
		for _, m := range mappings[clientID.String()] {
			key := m.AWSAccountID + "|" + m.AWSRoleARN
			if _, seen := allowed[key]; seen {
				continue
			}

			roleName := roleNameFromARN(m.AWSRoleARN)
			grant := Grant{
				AccountID:    m.AWSAccountID,
				AccountAlias: m.AWSAccountAlias,
				RoleARN:      m.AWSRoleARN,
				RoleName:     roleName,
			}
			allowed[key] = grant

			acct, ok := byAccount[m.AWSAccountID]
			if !ok {
				acct = &AccountAccess{AccountID: m.AWSAccountID, AccountAlias: m.AWSAccountAlias}
				byAccount[m.AWSAccountID] = acct
			}
			acct.Roles = append(acct.Roles, RoleAccess{RoleARN: m.AWSRoleARN, RoleName: roleName})
		}
	}

	accounts := make([]AccountAccess, 0, len(byAccount))
	for _, acct := range byAccount {
		sort.Slice(acct.Roles, func(i, j int) bool { return acct.Roles[i].RoleName < acct.Roles[j].RoleName })
		accounts = append(accounts, *acct)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Label() < accounts[j].Label() })

	return &Entitlements{Accounts: accounts, allowed: allowed}, nil
}

// roleNameFromARN extracts the role name from a role ARN, mirroring how the config server
// derives it. It falls back to the raw ARN if parsing fails.
func roleNameFromARN(roleARN string) string {
	parsed, err := arn.Parse(roleARN)
	if err != nil {
		return roleARN
	}
	return strings.TrimPrefix(parsed.Resource, "role/")
}
