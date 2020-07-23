package aws_config_server

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/honeycombio/beeline-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type oidcFederatedRoles struct {
	roles map[okta.ClientID][]accountAndRole
}

func (o *oidcFederatedRoles) Add(clientID okta.ClientID, accountAndRoles ...accountAndRole) {
	if o.roles == nil {
		o.roles = map[okta.ClientID][]accountAndRole{}
	}
	existing, ok := o.roles[clientID]
	if !ok {
		o.roles[clientID] = accountAndRoles
		return
	}
	o.roles[clientID] = append(existing, accountAndRoles...)
}

func (o *oidcFederatedRoles) Merge(others ...oidcFederatedRoles) {
	for _, other := range others {
		for clientID, accountAndRoles := range other.roles {
			o.Add(clientID, accountAndRoles...)
		}
	}
}

func getOIDCFederatedRoles(
	ctx context.Context,
	oidcProvider string,
	accountAndRoles []accountAndRole,
) (*oidcFederatedRoles, error) {

	oidcFederatedRoles := &oidcFederatedRoles{}
	identityProviderURL, err := url.Parse(oidcProvider)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse OIDC Provider input as an URL")
	}
	oidcProviderHostname := identityProviderURL.Hostname()

	for _, ar := range accountAndRoles {
		role := ar.Role
		if role == nil {
			logrus.Debug("nil role")
			continue
		}

		if role.AssumeRolePolicyDocument == nil {
			continue // role doesn't have an assume role policy document
		}
		policyDoc, err := NewPolicyDocument(*role.AssumeRolePolicyDocument)
		if err != nil {
			return nil, err
		}

		for _, statement := range policyDoc.Statements {
			clientIDs := statement.GetFederatedClientIDs(oidcProviderHostname)
			for _, clientID := range clientIDs {
				oidcFederatedRoles.Add(clientID, ar)
			}
		}
	}
	return oidcFederatedRoles, nil
}

func shouldSkipTags(tags []*iam.Tag) bool {
	for _, tag := range tags {
		if tag != nil && tag.Key != nil && *tag.Key == skipRolesTagKey {
			return true
		}
	}
	return false
}

// We can skip over roles with specific tags
func filterOIDCFederatedRoles(
	ctx context.Context,
	svc iamiface.IAMAPI,
	federatedRoles *oidcFederatedRoles,
) (*oidcFederatedRoles, error) {
	ctx, span := beeline.StartSpan(ctx, "filtering AWS roles")
	defer span.Send()

	filtered := &oidcFederatedRoles{}
	for clientID, accountAndRoles := range federatedRoles.roles {
		for _, accountAndRole := range accountAndRoles {
			skip, err := skipOIDCFederatedRole(ctx, svc, accountAndRole)
			if err != nil {
				return nil, err
			}
			if skip {
				continue
			}

			filtered.Add(clientID, accountAndRole)
		}
	}
	return filtered, nil
}

func skipOIDCFederatedRole(
	ctx context.Context,
	svc iamiface.IAMAPI,
	accountAndRole accountAndRole,
) (bool, error) {
	tags, err := listRoleTags(ctx, svc, accountAndRole.Role.RoleName)
	if err != nil {
		return true, err
	}
	return shouldSkipTags(tags), nil
}
