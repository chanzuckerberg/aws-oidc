// Package awsaccess is the single source of truth for the question "what AWS access does
// this person have". The config server projects it into the AWS config it serves, and the
// portal projects it into the entitlements it shows and the ceiling it enforces on agent
// grants. Keeping one traversal here means those two views can never disagree about a
// person's accounts and roles.
package awsaccess

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

// MappingsProvider returns the current rolemap grouped by Okta client ID. It is called once
// per request so callers always see the latest rolemap without a restart. The ConfigMap-backed
// implementation lives in pkg/configmap so this package stays free of the Kubernetes client.
type MappingsProvider func(ctx context.Context) (okta.OIDCRoleMappingsByKey, error)

// Role is a single role a person can assume in an account. OktaClientID is the app the role
// is reached through, which the config server needs for the credential_process client id.
type Role struct {
	OktaClientID string
	RoleARN      string
	RoleName     string
}

// Account is the set of roles a person can assume in one account.
type Account struct {
	ID    string
	Alias string
	Roles []Role
}

// Label is the account's display name, preferring the alias and falling back to the id.
func (a Account) Label() string {
	if a.Alias != "" {
		return a.Alias
	}
	return a.ID
}

// Access is a person's full grantable AWS access: their accounts and the roles in each. It
// is the ceiling for anything the portal lets them grant an agent.
type Access struct {
	Accounts []Account
}

// Resolve computes a person's access. It lists the Okta apps assigned to the subject (their
// client IDs) and looks each up in the rolemap, then deduplicates and sorts the result.
func Resolve(ctx context.Context, sub string, apps okta.AppLister, mappings okta.OIDCRoleMappingsByKey) (*Access, error) {
	clientIDs, err := okta.GetClientIDs(ctx, sub, apps)
	if err != nil {
		return nil, fmt.Errorf("getting client IDs for %s: %w", sub, err)
	}
	return FromClientIDs(clientIDs, mappings), nil
}

// FromClientIDs builds Access from an already-resolved set of client IDs. A person can reach
// the same role in the same account through more than one Okta app, so grants are
// deduplicated by account plus role ARN, keeping the first-seen mapping (and its client id).
func FromClientIDs(clientIDs []okta.ClientID, mappings okta.OIDCRoleMappingsByKey) *Access {
	byAccount := map[string]*Account{}
	seen := map[string]bool{}

	for _, clientID := range clientIDs {
		for _, m := range mappings[clientID.String()] {
			key := m.AWSAccountID + "|" + m.AWSRoleARN
			if seen[key] {
				continue
			}
			seen[key] = true

			acct, ok := byAccount[m.AWSAccountID]
			if !ok {
				acct = &Account{ID: m.AWSAccountID, Alias: m.AWSAccountAlias}
				byAccount[m.AWSAccountID] = acct
			}
			acct.Roles = append(acct.Roles, Role{
				OktaClientID: m.OktaClientID,
				RoleARN:      m.AWSRoleARN,
				RoleName:     RoleNameFromARN(m.AWSRoleARN),
			})
		}
	}

	accounts := make([]Account, 0, len(byAccount))
	for _, acct := range byAccount {
		sort.Slice(acct.Roles, func(i, j int) bool { return acct.Roles[i].RoleName < acct.Roles[j].RoleName })
		accounts = append(accounts, *acct)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].Label() < accounts[j].Label() })

	return &Access{Accounts: accounts}
}

// RoleNameFromARN extracts the role name from a role ARN, keeping any path so a role like
// arn:aws:iam::111:role/agents/data-bot-ro yields "agents/data-bot-ro". It falls back to the
// raw ARN if parsing fails.
func RoleNameFromARN(roleARN string) string {
	parsed, err := arn.Parse(roleARN)
	if err != nil {
		return roleARN
	}
	return strings.TrimPrefix(parsed.Resource, "role/")
}
