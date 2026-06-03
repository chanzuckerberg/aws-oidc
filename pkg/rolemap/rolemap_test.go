package rolemap

import (
	"reflect"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

func TestCustomRoleMappings(t *testing.T) {
	const (
		accountID    = "123456789012"
		accountAlias = "sci-data-dev"
	)

	cases := map[string]struct {
		value interface{}
		want  okta.OIDCRoleMappings
	}{
		"single role single client": {
			value: []interface{}{
				map[string]interface{}{
					"role_name":  "sci-data-dev-omero-access",
					"client_ids": []interface{}{"0oaCLIENT1"},
				},
			},
			want: okta.OIDCRoleMappings{
				{
					OktaClientID:    "0oaCLIENT1",
					AWSAccountID:    accountID,
					AWSAccountAlias: accountAlias,
					AWSRoleARN:      "arn:aws:iam::123456789012:role/sci-data-dev-omero-access",
				},
			},
		},
		"multiple roles and client ids": {
			value: []interface{}{
				map[string]interface{}{
					"role_name":  "role-a",
					"client_ids": []interface{}{"0oaA1", "0oaA2"},
				},
				map[string]interface{}{
					"role_name":  "role-b",
					"client_ids": []interface{}{"0oaB1"},
				},
			},
			want: okta.OIDCRoleMappings{
				{OktaClientID: "0oaA1", AWSAccountID: accountID, AWSAccountAlias: accountAlias, AWSRoleARN: "arn:aws:iam::123456789012:role/role-a"},
				{OktaClientID: "0oaA2", AWSAccountID: accountID, AWSAccountAlias: accountAlias, AWSRoleARN: "arn:aws:iam::123456789012:role/role-a"},
				{OktaClientID: "0oaB1", AWSAccountID: accountID, AWSAccountAlias: accountAlias, AWSRoleARN: "arn:aws:iam::123456789012:role/role-b"},
			},
		},
		// Accounts without custom roles omit the output; a nil value must yield nothing, not an error.
		"nil value yields empty": {
			value: nil,
			want:  nil,
		},
		"wrong top-level type yields empty": {
			value: "not-a-list",
			want:  nil,
		},
		"malformed entries are skipped": {
			value: []interface{}{
				"not-an-object",
				map[string]interface{}{"role_name": "", "client_ids": []interface{}{"x"}}, // empty name
				map[string]interface{}{"role_name": "no-clients"},                         // missing client_ids
				map[string]interface{}{
					"role_name":  "good-role",
					"client_ids": []interface{}{"0oaGOOD", 42}, // non-string client id skipped
				},
			},
			want: okta.OIDCRoleMappings{
				{OktaClientID: "0oaGOOD", AWSAccountID: accountID, AWSAccountAlias: accountAlias, AWSRoleARN: "arn:aws:iam::123456789012:role/good-role"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := customRoleMappings(accountID, accountAlias, tc.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Treat empty and nil slices as equivalent.
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got  %#v\nwant %#v", got, tc.want)
			}
		})
	}
}
