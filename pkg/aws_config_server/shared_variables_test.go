package aws_config_server

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	oidc "github.com/coreos/go-oidc"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

const oidcProvider = "https://localhost"

type failingVerifier struct{}

func (fv *failingVerifier) Verify(ctx context.Context, idToken string) (*oidc.IDToken, error) {
	return nil, fmt.Errorf("Failing verifier")
}

type emptyOktaApplications struct{}

func (oktaApp *emptyOktaApplications) ListApplications(ctx context.Context, qp *query.Params) ([]okta.App, *okta.Response, error) {
	return nil, nil, nil
}

type idTokenVerifier struct {
	expectedIDToken string
}

func (idtv *idTokenVerifier) Verify(ctx context.Context, idToken string) (*oidc.IDToken, error) {
	if idtv.expectedIDToken != idToken {
		return nil, fmt.Errorf("id tokens do not match!")
	}
	return &oidc.IDToken{}, nil
}

var testAWSConfigGenerationParams = AWSConfigGenerationParams{
	OIDCProvider:   "validProvider",
	AWSWorkerRole:  "validWorker",
	AWSMasterRoles: []string{"arn:aws:iam::AccountNumber1:role/MasterRole1"},
}

var samplePolicyDocument = &PolicyDocument{
	Statements: []StatementEntry{
		// All conditions met
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRoleWithWebIdentity"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: map[string]string{"localhost:aud": "clientIDValue1"},
			},
		},
		// Statement that doesn't meet all the qualifications
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRoleWithWebIdentity"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN", // no OIDC provider
			},
			Condition: Condition{
				StringEquals: map[string]string{"someID": "someResource"},
			},
		},
	},
}
var invalidPolicyStatements = &PolicyDocument{
	Statements: []StatementEntry{
		// Wrong Action
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRole"}, // where we check for action
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: map[string]string{"SAML:aud": "invalidClientID"},
			},
		},
		// Not federated (from Principal map)
		{
			Effect:    "Allow",
			Action:    []string{"sts:AssumeRoleWithWebIdentity"},
			Sid:       "",
			Principal: Principal{}, // where we check for Federation
			Condition: Condition{
				StringEquals: map[string]string{"SAML:aud": "invalidClientID"},
			},
		},
		// wrong provider
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRoleWithWebIdentity"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/anotherprovider", // where we check for provider
			},
			Condition: Condition{
				StringEquals: map[string]string{"anotherprovider:aud": "invalidClientID"},
			},
		},
		// invalid Structure for obtaining ClientKey
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRoleWithWebIdentity"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{},
		},
	},
}
var validPolicyStatements = []StatementEntry{
	// All conditions met with same clientID
	{
		Effect: "Allow",
		Action: []string{"sts:AssumeRoleWithWebIdentity"},
		Sid:    "",
		Principal: Principal{
			Federated: "ARN/localhost",
		},
		Condition: Condition{
			StringEquals: map[string]string{"localhost:aud": "clientIDValue1"},
		},
	},
	// All conditions met with another unique clientID
	{
		Effect: "Allow",
		Action: []string{"sts:AssumeRoleWithWebIdentity"},
		Sid:    "",
		Principal: Principal{
			Federated: "ARN/localhost",
		},
		Condition: Condition{
			StringEquals: map[string]string{"localhost:aud": "clientIDValue2"},
		},
	},
}

var revisedPolicyDocument = &PolicyDocument{
	Statements: []StatementEntry{
		// All conditions met (sts.AssumeRoleWithWebIdentity part of Action list)
		{
			Effect: "Allow",
			Action: []string{"sts:AssumeRoleWithWebIdentity", "sts:AssumeRoleWithSAML"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: map[string]string{"localhost:aud": "clientIDValue3"},
			},
		},
		// Invalid statement with revised StatementEntry
		{
			Effect: "Allow",
			Action: []string{"sts:InvalidAction"},
			Sid:    "",
			Principal: Principal{
				Federated: "ARN/localhost",
			},
			Condition: Condition{
				StringEquals: map[string]string{"localhost:aud": "invalidClientID"},
			},
		},
	},
}

// Custom structs needed for scenario with single actions encoded as a single string
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
				StringEquals: map[string]string{"localhost:aud": "clientIDValue4"},
			},
		},
	},
}

var testRoles0 = []*iam.Role{
	{
		Arn:      aws.String("roleARN"),
		RoleName: aws.String("testRoles0"),
	},
}
var testRoles1 = []*iam.Role{
	{
		Arn:      aws.String("roleARN0"),
		RoleName: aws.String("testRoles1"),
	},
}
var testRoles2 = []*iam.Role{
	{
		Arn:      aws.String("roleARN1"),
		RoleName: aws.String("testRoles2"),
	},
}

var testRoles3 = []*iam.Role{
	{
		Arn:      aws.String("roleARN0"),
		RoleName: aws.String("testRoles3"),
	},
}

var testConfigMapping = map[string][]ConfigProfile{
	"clientID1": {
		{
			acctName: "Account1",
			roleARN:  "arn:aws:iam::AccountNumber1:role/WorkerRole",
		},
		{
			acctName: "Account2",
			roleARN:  "arn:aws:iam::AccountNumber2:role/WorkerRole",
		},
	},
	"clientID2": {
		{
			acctName: "Account3",
			roleARN:  "arn:aws:iam::AccountNumber3:role/WorkerRole",
		},
	},
	"clientID3": {
		{
			acctName: "account with space",
			roleARN:  "arn:aws:iam::account-with-space:role/WorkerRole",
		},
	},
}
