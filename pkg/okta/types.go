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
type OIDCRoleMappingByClientID []OIDCRoleMapping

func FromContext(ctx context.Context) *OIDCRoleMappingByClientID {
	v, _ := ctx.Value(ctxKey{}).(*OIDCRoleMappingByClientID)
	return v
}

func NewContext(parent context.Context, v *OIDCRoleMappingByClientID) context.Context {
	return context.WithValue(parent, ctxKey{}, v)
}
