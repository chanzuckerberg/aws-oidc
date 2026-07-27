package portal

import (
	"testing"

	"github.com/stretchr/testify/require"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

func TestAgentToCRRoundTrip(t *testing.T) {
	in := Agent{
		Name:       "data-bot",
		Owner:      "sub-1",
		OwnerEmail: "a@example.com",
		Grants: []Grant{
			{AccountID: "111111111111", AccountAlias: "prod", RoleARN: "arn:aws:iam::111111111111:role/agents/data-bot-ro", RoleName: "agents/data-bot-ro"},
		},
	}

	cr := agentToCR(in, "test-ns")
	require.Equal(t, "data-bot", cr.Name)
	require.Equal(t, "test-ns", cr.Namespace)
	require.Equal(t, "sub-1", cr.Spec.Owner)
	require.Len(t, cr.Spec.Grants, 1)
	require.NotNil(t, cr.Spec.Grants[0].AWS)
	require.Equal(t, "111111111111", cr.Spec.Grants[0].AWS.AccountID)
	require.Equal(t, "arn:aws:iam::111111111111:role/agents/data-bot-ro", cr.Spec.Grants[0].AWS.RoleARN)

	out := agentFromCR(cr)
	require.Equal(t, in.Name, out.Name)
	require.Equal(t, in.Owner, out.Owner)
	require.Equal(t, in.OwnerEmail, out.OwnerEmail)
	require.Equal(t, in.Grants, out.Grants)
}

// A grant with no known provider section is skipped when mapping back to the portal, which
// only renders AWS.
func TestAgentFromCRSkipsNonAWSGrants(t *testing.T) {
	cr := &agentsv1.Agent{}
	cr.Name = "bot"
	cr.Spec.Owner = "sub-1"
	cr.Spec.Grants = []agentsv1.Grant{
		{AWS: &agentsv1.AWSGrant{AccountID: "111111111111", RoleARN: "arn:aws:iam::111111111111:role/x"}},
		{}, // future provider the portal does not render yet
	}

	out := agentFromCR(cr)
	require.Len(t, out.Grants, 1)
	require.Equal(t, "111111111111", out.Grants[0].AccountID)
}

func TestAgentToUnstructuredRoundTrip(t *testing.T) {
	cr := agentToCR(Agent{Name: "bot", Owner: "sub-1", Grants: []Grant{
		{AccountID: "111111111111", RoleARN: "arn:aws:iam::111111111111:role/x", RoleName: "x"},
	}}, "ns")

	u, err := agentToUnstructured(cr)
	require.NoError(t, err)
	require.Equal(t, "Agent", u.GetKind())
	require.Equal(t, "bot", u.GetName())

	back, err := agentFromUnstructured(u)
	require.NoError(t, err)
	require.Equal(t, cr.Spec, back.Spec)
}
