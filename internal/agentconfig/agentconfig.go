// Package agentconfig projects Agent custom resources into the config server's response
// shape. It pairs each agent's desired grants (spec) with what the operator provisioned
// (status) and emits one profile per provisioned grant, carrying the shared agent Okta
// client id and the per-agent role ARN. It is the config-server counterpart to the human
// projection in pkg/aws_config_server.
package agentconfig

import (
	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
)

// Build turns the agents a person owns into their config-server representation. Every agent
// profile carries the shared agentClientID as its client id, which is what scopes these
// profiles to agents through the role trust conditions. Agents with no provisioned grants
// are omitted so a not-yet-reconciled agent does not appear as an empty config.
func Build(agents []agentsv1.Agent, agentClientID, issuerURL string) []server.AgentConfig {
	out := make([]server.AgentConfig, 0, len(agents))
	for i := range agents {
		profiles := profilesForAgent(&agents[i], agentClientID, issuerURL)
		if len(profiles) == 0 {
			continue
		}
		out = append(out, server.AgentConfig{Name: agents[i].Name, Profiles: profiles})
	}
	return out
}

// profilesForAgent pairs spec and status grants by index (the reconciler writes status
// aligned with spec) and emits a profile only for grants the operator has provisioned. The
// account id and alias come from spec, since status carries only the provisioned role ARN.
func profilesForAgent(agent *agentsv1.Agent, agentClientID, issuerURL string) []server.AWSProfile {
	profiles := make([]server.AWSProfile, 0, len(agent.Spec.Grants))
	for i := range agent.Spec.Grants {
		specGrant := agent.Spec.Grants[i].AWS
		if specGrant == nil {
			continue
		}
		if i >= len(agent.Status.Grants) {
			continue
		}
		st := agent.Status.Grants[i]
		if st.State != agentsv1.GrantStateProvisioned || st.AWS == nil || st.AWS.RoleARN == "" {
			continue
		}

		roleName := specGrant.RoleName
		if roleName == "" {
			roleName = awsaccess.RoleNameFromARN(st.AWS.RoleARN)
		}
		profiles = append(profiles, server.AWSProfile{
			ClientID: agentClientID,
			RoleARN:  st.AWS.RoleARN,
			AWSAccount: server.AWSAccount{
				ID:    specGrant.AccountID,
				Name:  specGrant.AccountAlias,
				Alias: specGrant.AccountAlias,
			},
			IssuerURL: issuerURL,
			RoleName:  roleName,
		})
	}
	return profiles
}
