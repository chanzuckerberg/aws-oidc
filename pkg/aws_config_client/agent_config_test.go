package aws_config_client

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
)

func awsProfile(alias, roleARN, roleName string) server.AWSProfile {
	return server.AWSProfile{
		ClientID:   "0oaAGENT",
		RoleARN:    roleARN,
		RoleName:   roleName,
		IssuerURL:  "https://czi.okta.com",
		AWSAccount: server.AWSAccount{ID: "111111111111", Name: alias, Alias: alias},
	}
}

func TestAgentConfigWrite(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	renderer := NewAgentConfigRenderer(agentsDir, DefaultAWSRegion)

	agents := []server.AgentConfig{{
		Name: "data-bot",
		Profiles: []server.AWSProfile{
			awsProfile("prod", "arn:aws:iam::111111111111:role/agents/data-bot-ro", "agents/data-bot-ro"),
			awsProfile("dev", "arn:aws:iam::222222222222:role/agents/data-bot-rw", "agents/data-bot-rw"),
		},
	}}

	err := renderer.Write(agents)
	require.NoError(t, err)

	configPath := filepath.Join(agentsDir, "data-bot", "config")
	f, err := ini.Load(configPath)
	require.NoError(t, err)

	// Profile names are <account>-<role>, sanitized (the pathed role loses its slash). The
	// agent name is not in the profile name.
	require.NotNil(t, f.Section("profile prod-agents-data-bot-ro"))
	require.True(t, f.HasSection("profile prod-agents-data-bot-ro"))
	require.True(t, f.HasSection("profile dev-agents-data-bot-rw"))

	// The first grant is also the agent-scoped default.
	require.True(t, f.HasSection("profile "+AgentScopedProfile))
	scoped := f.Section("profile " + AgentScopedProfile).Key(AWSConfigSectionCredentialProcess).String()
	first := f.Section("profile prod-agents-data-bot-ro").Key(AWSConfigSectionCredentialProcess).String()
	require.Equal(t, first, scoped)

	require.NotContains(t, first, "--node-local-cache")
	require.Contains(t, first, "--aws-role-arn=arn:aws:iam::111111111111:role/agents/data-bot-ro")
	require.Contains(t, first, "--client-id=0oaAGENT")

	require.Equal(t, DefaultAWSRegion, f.Section("profile prod-agents-data-bot-ro").Key(AWSConfigSectionRegion).String())
}

func TestAgentConfigPrunesStaleAgents(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	// A previously configured agent the person no longer owns.
	staleDir := filepath.Join(agentsDir, "old-bot")
	require.NoError(t, os.MkdirAll(staleDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(staleDir, "config"), []byte("stale"), 0600))

	renderer := NewAgentConfigRenderer(agentsDir, DefaultAWSRegion)
	agents := []server.AgentConfig{{
		Name:     "data-bot",
		Profiles: []server.AWSProfile{awsProfile("prod", "arn:aws:iam::111111111111:role/agents/x", "agents/x")},
	}}

	require.NoError(t, renderer.Write(agents))

	_, err := os.Stat(staleDir)
	require.True(t, os.IsNotExist(err), "the stale agent directory should be pruned")
	require.DirExists(t, filepath.Join(agentsDir, "data-bot"))
}

func TestAgentConfigEmptyListPrunesAll(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	require.NoError(t, os.MkdirAll(filepath.Join(agentsDir, "a"), 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(agentsDir, "b"), 0700))

	renderer := NewAgentConfigRenderer(agentsDir, DefaultAWSRegion)
	// A present-but-empty list (every agent revoked) prunes everything.
	require.NoError(t, renderer.Write([]server.AgentConfig{}))

	entries, err := os.ReadDir(agentsDir)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestAgentConfigWriteNilCreatesNothing(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	renderer := NewAgentConfigRenderer(agentsDir, DefaultAWSRegion)
	require.NoError(t, renderer.Write(nil))

	_, err := os.Stat(agentsDir)
	require.True(t, os.IsNotExist(err), "no agents dir should be created when there are no agents")
}

func TestAgentConfigPrint(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	renderer := NewAgentConfigRenderer(agentsDir, DefaultAWSRegion)

	agents := []server.AgentConfig{{
		Name:     "data-bot",
		Profiles: []server.AWSProfile{awsProfile("prod", "arn:aws:iam::111111111111:role/agents/x", "agents/x")},
	}}

	var buf bytes.Buffer
	require.NoError(t, renderer.Print(agents, &buf))

	out := buf.String()
	require.Contains(t, out, filepath.Join(agentsDir, "data-bot", "config"))
	require.Contains(t, out, "profile prod-agents-x")

	// Print must not touch disk.
	_, err := os.Stat(agentsDir)
	require.True(t, os.IsNotExist(err))
}
