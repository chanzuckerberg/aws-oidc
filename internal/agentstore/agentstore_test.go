package agentstore

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

func TestUnstructuredRoundTrip(t *testing.T) {
	cr := &agentsv1.Agent{
		TypeMeta:   metav1.TypeMeta{APIVersion: agentsv1.GroupVersion.String(), Kind: "Agent"},
		ObjectMeta: metav1.ObjectMeta{Name: "data-bot", Namespace: "ns"},
		Spec: agentsv1.AgentSpec{
			Owner:      "sub-1",
			OwnerEmail: "a@example.com",
			Grants: []agentsv1.Grant{
				{AWS: &agentsv1.AWSGrant{
					AccountID:    "111111111111",
					AccountAlias: "prod",
					RoleARN:      "arn:aws:iam::111111111111:role/agents/data-bot-ro",
					RoleName:     "agents/data-bot-ro",
				}},
			},
		},
		Status: agentsv1.AgentStatus{
			Grants: []agentsv1.GrantStatus{{
				Provider: agentsv1.ProviderAWS,
				State:    agentsv1.GrantStateProvisioned,
				AWS:      &agentsv1.AWSGrantStatus{AccountID: "111111111111", RoleARN: "arn:aws:iam::111111111111:role/agents/data-bot-ro"},
			}},
		},
	}

	u, err := ToUnstructured(cr)
	require.NoError(t, err)
	require.Equal(t, "Agent", u.GetKind())
	require.Equal(t, "data-bot", u.GetName())

	back, err := FromUnstructured(u)
	require.NoError(t, err)
	require.Equal(t, cr.Spec, back.Spec)
	require.Equal(t, cr.Status, back.Status)
}
