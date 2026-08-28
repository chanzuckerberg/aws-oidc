package agentpod

import (
	"fmt"
	"strings"

	"gopkg.in/ini.v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/internal/agentconfig"
	awsconfigclient "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_client"
)

// renderAWSConfig builds the AWS config file the agent's pods mount. It is the in-cluster
// counterpart to the laptop rendering in pkg/aws_config_client: same profile names, but the
// credentials come from the pod's projected service account token exchanged with STS rather
// than from a credential_process running aws-oidc. Profile names match so a prompt or script
// that works on a laptop works in a pod.
//
// It returns an empty string when the agent has no provisioned grants, which is how a pod
// waiting on its roles is distinguished from one that has them.
func (r *Reconciler) renderAWSConfig(agent *agentsv1.Agent) (string, error) {
	// The Okta client id and issuer only matter to the laptop rendering, which mints tokens
	// through Okta. In the cluster the token comes from Kubernetes, so they are left empty.
	configs := agentconfig.Build([]agentsv1.Agent{*agent}, "", "")
	if len(configs) == 0 {
		return "", nil
	}
	profiles := configs[0].Profiles

	out := ini.Empty()
	for i := range profiles {
		profile := profiles[i]

		err := r.addProfile(out, awsconfigclient.AgentProfileName(profile), profile.RoleARN, agent)
		if err != nil {
			return "", err
		}

		// The first grant doubles as the default profile so a bare `aws` call works, matching
		// what the laptop rendering does.
		if i == 0 {
			err = r.addProfile(out, awsconfigclient.AgentScopedProfile, profile.RoleARN, agent)
			if err != nil {
				return "", err
			}
		}
	}

	var rendered strings.Builder
	_, err := out.WriteTo(&rendered)
	if err != nil {
		return "", fmt.Errorf("rendering aws config: %w", err)
	}
	return rendered.String(), nil
}

func (r *Reconciler) addProfile(out *ini.File, name, roleARN string, agent *agentsv1.Agent) error {
	section, err := out.NewSection("profile " + name)
	if err != nil {
		return fmt.Errorf("creating profile %s: %w", name, err)
	}
	section.Key(awsconfigclient.AWSConfigSectionOutput).SetValue("json")
	section.Key(awsconfigclient.AWSConfigSectionRegion).SetValue(r.Region)
	section.Key("role_arn").SetValue(roleARN)
	section.Key("web_identity_token_file").SetValue(tokenFilePath)
	section.Key("role_session_name").SetValue(sessionName(agent))
	return nil
}

// sessionName builds the STS session name for the agent. It encodes the agent name and
// owner email so CloudTrail entries are attributable without inspecting the token subject.
// STS caps session names at 64 characters; the value is truncated if needed.
func sessionName(agent *agentsv1.Agent) string {
	name := "agent-" + agent.Name
	if agent.Spec.OwnerEmail != "" {
		name = name + "/" + agent.Spec.OwnerEmail
	}
	if len(name) > 64 {
		return name[:64]
	}
	return name
}
