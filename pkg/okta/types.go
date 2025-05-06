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

func FromContext(ctx context.Context) *OIDCRoleMappings {
	v, _ := ctx.Value(ctxKey{}).(*OIDCRoleMappings)
	return v
}

func NewContext(parent context.Context, v *OIDCRoleMappings) context.Context {
	return context.WithValue(parent, ctxKey{}, v)
}
