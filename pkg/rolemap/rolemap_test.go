package rolemap

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

func TestGenerateFromS3State(t *testing.T) {
	state := `{
		"version": 4,
		"outputs": {
			"current_account_id": {"value": "123456789012", "type": "string"},
			"current_account_alias": {"value": "sci-data-dev", "type": "string"},
			"czi-okta_okta_czi_admin_oidc_client_ids": {"value": ["admin-client"], "type": ["list", "string"]},
			"czi_okta_poweruser_oidc_client_ids": {"value": ["poweruser-client"], "type": ["list", "string"]},
			"czi_okta_readonly_oidc_client_ids": {"value": ["readonly-client"], "type": ["list", "string"]},
			"czi_okta_custom_role_oidc_client_ids": {
				"value": [{"role_name": "custom-role", "client_ids": ["custom-client"]}],
				"type": ["list", ["object", {}]]
			}
		}
	}`
	client := &mockS3Client{
		listOutputs: []*s3.ListObjectsV2Output{
			{
				Contents: []types.Object{
					{Key: aws.String("terraform/shared-infra/accounts/es-prod.tfstate")},
					{Key: aws.String("terraform/shared-infra/accounts/sci-data-dev.tfstate")},
					{Key: aws.String("terraform/shared-infra/accounts/not-state.txt")},
				},
				IsTruncated:           aws.Bool(true),
				NextContinuationToken: aws.String("next"),
			},
			{
				Contents:    []types.Object{},
				IsTruncated: aws.Bool(false),
			},
		},
		objects: map[string]string{
			"terraform/shared-infra/accounts/sci-data-dev.tfstate": state,
		},
	}

	got, err := Generate(context.Background(), client, "terragrunt-engine-state", "terraform/shared-infra/accounts/")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	want := okta.OIDCRoleMappings{
		{OktaClientID: "readonly-client", AWSAccountID: "123456789012", AWSAccountAlias: "sci-data-dev", AWSRoleARN: "arn:aws:iam::123456789012:role/readonly"},
		{OktaClientID: "poweruser-client", AWSAccountID: "123456789012", AWSAccountAlias: "sci-data-dev", AWSRoleARN: "arn:aws:iam::123456789012:role/poweruser"},
		{OktaClientID: "custom-client", AWSAccountID: "123456789012", AWSAccountAlias: "sci-data-dev", AWSRoleARN: "arn:aws:iam::123456789012:role/custom-role"},
		{OktaClientID: "admin-client", AWSAccountID: "123456789012", AWSAccountAlias: "sci-data-dev", AWSRoleARN: "arn:aws:iam::123456789012:role/okta-czi-admin"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Generate() = %#v, want %#v", got, want)
	}
	if len(client.listInputs) != 2 || aws.ToString(client.listInputs[1].ContinuationToken) != "next" {
		t.Errorf("ListObjectsV2 pagination inputs = %#v", client.listInputs)
	}
}

func TestStateRoleMappingsSkipsStateWithoutRoleOutputs(t *testing.T) {
	got, err := stateRoleMappings(strings.NewReader(`{"outputs": {}}`))
	if err != nil {
		t.Fatalf("stateRoleMappings() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("stateRoleMappings() = %#v, want no mappings", got)
	}
}

func TestStateRoleMappingsRequiresAccountOutputsForRoleState(t *testing.T) {
	state := `{"outputs": {"czi_okta_poweruser_oidc_client_ids": {"value": ["client"]}}}`
	_, err := stateRoleMappings(strings.NewReader(state))
	if err == nil || !strings.Contains(err.Error(), _accountIDOutputName) {
		t.Fatalf("stateRoleMappings() error = %v, want missing account ID output", err)
	}
}

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

type mockS3Client struct {
	listInputs  []*s3.ListObjectsV2Input
	listOutputs []*s3.ListObjectsV2Output
	objects     map[string]string
}

func (m *mockS3Client) ListObjectsV2(
	_ context.Context,
	input *s3.ListObjectsV2Input,
	_ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	m.listInputs = append(m.listInputs, input)
	output := m.listOutputs[0]
	m.listOutputs = m.listOutputs[1:]
	return output, nil
}

func (m *mockS3Client) GetObject(
	_ context.Context,
	input *s3.GetObjectInput,
	_ ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	value, ok := m.objects[aws.ToString(input.Key)]
	if !ok {
		return nil, io.EOF
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewBufferString(value))}, nil
}
