package okta

import "context"

type ClientID string

func (c ClientID) String() string {
	return string(c)
}

type OIDCRoleMapping struct {
	AWSAccountID    string `yaml:"aws_account_id"`
	AWSAccountAlias string `yaml:"aws_account_alias"`
	AWSRoleARN      string `yaml:"aws_role_arn"`
	OktaClientID    string `yaml:"okta_client_id"`
}

type ctxKey struct{}
type OIDCRoleMappings []OIDCRoleMapping
type OIDCRoleMappingsByKey map[string][]OIDCRoleMapping

// ByClientID groups the mappings by their Okta client ID, the shape the config
// server looks up per request.
func (m OIDCRoleMappings) ByClientID() OIDCRoleMappingsByKey {
	byKey := make(OIDCRoleMappingsByKey, len(m))
	for _, mapping := range m {
		byKey[mapping.OktaClientID] = append(byKey[mapping.OktaClientID], mapping)
	}
	return byKey
}

func FromContext(ctx context.Context) *OIDCRoleMappings {
	v, _ := ctx.Value(ctxKey{}).(*OIDCRoleMappings)
	return v
}

func NewContext(parent context.Context, v *OIDCRoleMappings) context.Context {
	return context.WithValue(parent, ctxKey{}, v)
}
