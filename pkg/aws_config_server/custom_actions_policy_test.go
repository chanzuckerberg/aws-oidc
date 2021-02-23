package aws_config_server

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	cziAWS "github.com/chanzuckerberg/go-misc/aws"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var testRoles3 = []*iam.Role{
	{
		Arn:                      aws.String(BareRoleARN("roleARN0").String()),
		RoleName:                 aws.String("testRoles3"),
		AssumeRolePolicyDocument: policyDocumentToString(revisedPolicyDocument),
	},
}

var revisedPolicyDocument = &PolicyDocument{
	Statements: []StatementEntry{
		// All conditions met (sts.AssumeRoleWithWebIdentity part of Action list)
		{
			Effect: "Allow",
			Action: Action{"sts:AssumeRoleWithWebIdentity", "sts:AssumeRoleWithSAML"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: StringEqualsCondition{
					"localhost:aud": []string{"clientIDValue3"},
				},
			},
		},
		// Invalid statement with revised StatementEntry
		{
			Effect: "Allow",
			Action: Action{"sts:InvalidAction"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: StringEqualsCondition{
					"localhost:aud": []string{"invalidClientID"},
				},
			},
		},
	},
}

func policyDocumentToString(policyDoc *PolicyDocument) *string {
	jsonPolicyData, err := json.Marshal(policyDoc)
	if err != nil {
		panic(err)
	}
	return aws.String(url.PathEscape(string(jsonPolicyData)))
}

func TestParseMultipleActions(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)
	ctrl := gomock.NewController(t)

	client := &cziAWS.Client{}
	_, mockIAM := client.WithMockIAM(ctrl)

	mockIAM.EXPECT().
		ListRolesPagesWithContext(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			ctx context.Context,
			input *iam.ListRolesInput,
			accumulatorFunc func(*iam.ListRolesOutput, bool) bool,
		) error {
			accumulatorFunc(&iam.ListRolesOutput{Roles: testRoles3}, true)
			return nil
		},
	).AnyTimes()

	iamOutput, err := listRoles(ctx, mockIAM)
	r.NoError(err)
	r.Len(iamOutput, 1)
}

// // Custom structs needed for scenario with single actions encoded as a single string
type AlternateStatementEntry struct {
	Effect    string    `json:"Effect"`
	Action    string    `json:"Action"`
	Sid       string    `json:"Sid"`
	Principal Principal `json:"Principal"`
	Condition Condition `json:"Condition"`
}
type AlternatePolicyDocument struct {
	Version    string                    `json:"Version"`
	Statements []AlternateStatementEntry `json:"Statement"`
}

var alternatePolicyDocument = &AlternatePolicyDocument{
	Statements: []AlternateStatementEntry{
		{
			Effect: "Allow",
			Action: "sts:AssumeRoleWithWebIdentity",
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: StringEqualsCondition(
					map[string][]string{"localhost:aud": {"clientIDValue4"}},
				),
			},
		},
	},
}

func TestSingleStringAction(t *testing.T) {
	r := require.New(t)

	policyData, err := json.Marshal(alternatePolicyDocument)
	r.NoError(err)

	policyDoc := PolicyDocument{}
	err = json.Unmarshal(policyData, &policyDoc)
	r.NoError(err)

	r.NotEmpty(policyDoc)
	r.Len(policyDoc.Statements, 1)
	r.Len(policyDoc.Statements[0].Action, 1)
	r.Equal(policyDoc.Statements[0].Action[0], "sts:AssumeRoleWithWebIdentity")
}
