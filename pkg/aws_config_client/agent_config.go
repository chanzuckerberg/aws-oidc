package aws_config_client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"gopkg.in/ini.v1"
)

// AgentScopedProfile is the profile name a bare `aws` command uses inside an agent sandbox.
// It matches AWS_PROFILE=agent-scoped in the agent's managed settings, so the agent's first
// grant works with no explicit profile.
const AgentScopedProfile = "agent-scoped"

// AgentConfigsDir is the base directory agent config files live under. It sits outside
// ~/.aws on purpose: the agent's managed settings deny reads under ~/.aws, and this path is
// user-writable so `configure` (which runs as the person) can create it.
func AgentConfigsDir(home string) string {
	return filepath.Join(home, ".aws-oidc", "agents")
}

// AgentConfigRenderer writes one AWS config file per agent under a base directory.
type AgentConfigRenderer struct {
	agentsDir string
	region    string
}

// NewAgentConfigRenderer returns a renderer writing under agentsDir, applying region to
// every profile.
func NewAgentConfigRenderer(agentsDir, region string) *AgentConfigRenderer {
	return &AgentConfigRenderer{agentsDir: agentsDir, region: region}
}

// Write regenerates one config file per agent at <agentsDir>/<name>/config, then prunes any
// agent directory not present in agents so revoked agents disappear rather than lingering
// empty. Each file is a full rewrite (no merge), since nothing else owns it. Passing an
// empty, non-nil slice prunes every agent directory.
func (r *AgentConfigRenderer) Write(agents []server.AgentConfig) error {
	for i := range agents {
		agent := agents[i]
		file, err := r.render(agent)
		if err != nil {
			return err
		}

		dir := filepath.Join(r.agentsDir, agent.Name)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return fmt.Errorf("creating agent dir %s: %w", dir, err)
		}

		writer := NewAWSConfigFileWriter(filepath.Join(dir, "config"))
		_, err = file.WriteTo(writer)
		if err != nil {
			return fmt.Errorf("rendering agent config for %s: %w", agent.Name, err)
		}
		err = writer.Finalize()
		if err != nil {
			return fmt.Errorf("writing agent config for %s: %w", agent.Name, err)
		}
	}
	return r.prune(agents)
}

// Print writes each agent's config to w with a path header, for --print-only.
func (r *AgentConfigRenderer) Print(agents []server.AgentConfig, w io.Writer) error {
	for i := range agents {
		agent := agents[i]
		file, err := r.render(agent)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\n# %s\n", filepath.Join(r.agentsDir, agent.Name, "config"))
		if err != nil {
			return err
		}
		_, err = file.WriteTo(w)
		if err != nil {
			return fmt.Errorf("printing agent config for %s: %w", agent.Name, err)
		}
	}
	return nil
}

func (r *AgentConfigRenderer) render(agent server.AgentConfig) (*ini.File, error) {
	out := ini.Empty()

	for i := range agent.Profiles {
		profile := agent.Profiles[i]
		err := r.addProfile(out, "profile "+AgentProfileName(profile), profile)
		if err != nil {
			return nil, err
		}
		// The first grant doubles as the agent-scoped default so a bare `aws` works.
		if i == 0 {
			err = r.addProfile(out, "profile "+AgentScopedProfile, profile)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (r *AgentConfigRenderer) addProfile(out *ini.File, section string, profile server.AWSProfile) error {
	credsProcess := fmt.Sprintf(
		"aws-oidc creds-process --issuer-url=%s --client-id=%s --aws-role-arn=%s",
		profile.IssuerURL,
		profile.ClientID,
		profile.RoleARN,
	)

	out.DeleteSection(section)
	sec, err := out.NewSection(section)
	if err != nil {
		return fmt.Errorf("creating %s section in agent config: %w", section, err)
	}
	sec.Key(AWSConfigSectionOutput).SetValue("json")
	sec.Key(AWSConfigSectionCredentialProcess).SetValue(credsProcess)
	sec.Key(AWSConfigSectionRegion).SetValue(r.region)
	return nil
}

// prune removes any agent directory not present in agents. A missing base directory is not
// an error.
func (r *AgentConfigRenderer) prune(agents []server.AgentConfig) error {
	entries, err := os.ReadDir(r.agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading agents dir %s: %w", r.agentsDir, err)
	}

	keep := make(map[string]bool, len(agents))
	for _, a := range agents {
		keep[a.Name] = true
	}

	for _, entry := range entries {
		if !entry.IsDir() || keep[entry.Name()] {
			continue
		}
		path := filepath.Join(r.agentsDir, entry.Name())
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("pruning stale agent dir %s: %w", path, err)
		}
	}
	return nil
}

// AgentProfileName is "<account>-<role>", sanitized the same way human account profile names
// are. The agent name is not included because it is already the directory. The operator uses
// it too, so an agent's profiles are named the same whether it runs on a laptop or in a pod.
func AgentProfileName(profile server.AWSProfile) string {
	return sanitizeProfileName(fmt.Sprintf("%s-%s", profile.AWSAccount.GetAliasOrName(), profile.RoleName))
}
