package agentconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
)

const (
	testClientID = "0oaAGENT"
	testIssuer   = "https://czi.okta.com"
)

func agent(name string, spec []agentsv1.Grant, status []agentsv1.GrantStatus) agentsv1.Agent {
	return agentsv1.Agent{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       agentsv1.AgentSpec{Owner: "sub-1", Grants: spec},
		Status:     agentsv1.AgentStatus{Grants: status},
	}
}

func awsSpec(accountID, alias, roleARN string) agentsv1.Grant {
	return agentsv1.Grant{AWS: &agentsv1.AWSGrant{AccountID: accountID, AccountAlias: alias, RoleARN: roleARN}}
}

func provisioned(roleARN string) agentsv1.GrantStatus {
	return agentsv1.GrantStatus{Provider: agentsv1.ProviderAWS, State: agentsv1.GrantStateProvisioned, AWS: &agentsv1.AWSGrantStatus{RoleARN: roleARN}}
}

func TestBuildProvisionedGrant(t *testing.T) {
	agents := []agentsv1.Agent{
		agent("data-bot",
			[]agentsv1.Grant{awsSpec("111111111111", "prod", "arn:aws:iam::111111111111:role/agents/data-bot-src")},
			[]agentsv1.GrantStatus{provisioned("arn:aws:iam::111111111111:role/agents/data-bot-ro")},
		),
	}

	got := Build(agents, testClientID, testIssuer)
	require.Len(t, got, 1)
	require.Equal(t, "data-bot", got[0].Name)
	require.Len(t, got[0].Profiles, 1)

	p := got[0].Profiles[0]
	// Client id is the shared agent app, not per-account.
	require.Equal(t, testClientID, p.ClientID)
	// Role ARN comes from status (the provisioned per-agent role), not spec.
	require.Equal(t, "arn:aws:iam::111111111111:role/agents/data-bot-ro", p.RoleARN)
	// Account comes from spec, which status does not carry.
	require.Equal(t, "111111111111", p.AWSAccount.ID)
	require.Equal(t, "prod", p.AWSAccount.Alias)
	require.Equal(t, testIssuer, p.IssuerURL)
	// Role name keeps the path.
	require.Equal(t, "agents/data-bot-ro", p.RoleName)
}

func TestBuildSkipsUnprovisionedGrants(t *testing.T) {
	agents := []agentsv1.Agent{
		agent("half-bot",
			[]agentsv1.Grant{
				awsSpec("111111111111", "prod", "arn:aws:iam::111111111111:role/src-a"),
				awsSpec("222222222222", "dev", "arn:aws:iam::222222222222:role/src-b"),
				awsSpec("333333333333", "stg", "arn:aws:iam::333333333333:role/src-c"),
			},
			[]agentsv1.GrantStatus{
				provisioned("arn:aws:iam::111111111111:role/agents/a"),
				{Provider: agentsv1.ProviderAWS, State: agentsv1.GrantStatePending},
				{Provider: agentsv1.ProviderAWS, State: agentsv1.GrantStateProvisioned, AWS: &agentsv1.AWSGrantStatus{RoleARN: ""}},
			},
		),
	}

	got := Build(agents, testClientID, testIssuer)
	require.Len(t, got, 1)
	// Only the first grant is provisioned with a role ARN.
	require.Len(t, got[0].Profiles, 1)
	require.Equal(t, "arn:aws:iam::111111111111:role/agents/a", got[0].Profiles[0].RoleARN)
}

func TestBuildPairsByIndex(t *testing.T) {
	// Status shorter than spec: the unpaired trailing spec grant is skipped.
	agents := []agentsv1.Agent{
		agent("bot",
			[]agentsv1.Grant{
				awsSpec("111111111111", "prod", "arn:aws:iam::111111111111:role/src-a"),
				awsSpec("222222222222", "dev", "arn:aws:iam::222222222222:role/src-b"),
			},
			[]agentsv1.GrantStatus{provisioned("arn:aws:iam::111111111111:role/agents/a")},
		),
	}

	got := Build(agents, testClientID, testIssuer)
	require.Len(t, got[0].Profiles, 1)
	require.Equal(t, "111111111111", got[0].Profiles[0].AWSAccount.ID)
}

func TestBuildOmitsAgentsWithNoProvisionedGrants(t *testing.T) {
	agents := []agentsv1.Agent{
		agent("pending-bot",
			[]agentsv1.Grant{awsSpec("111111111111", "prod", "arn:aws:iam::111111111111:role/src")},
			[]agentsv1.GrantStatus{{Provider: agentsv1.ProviderAWS, State: agentsv1.GrantStatePending}},
		),
		agent("ready-bot",
			[]agentsv1.Grant{awsSpec("222222222222", "dev", "arn:aws:iam::222222222222:role/src")},
			[]agentsv1.GrantStatus{provisioned("arn:aws:iam::222222222222:role/agents/x")},
		),
	}

	got := Build(agents, testClientID, testIssuer)
	require.Len(t, got, 1)
	require.Equal(t, "ready-bot", got[0].Name)
}

func TestBuildEmpty(t *testing.T) {
	require.Empty(t, Build(nil, testClientID, testIssuer))
}
