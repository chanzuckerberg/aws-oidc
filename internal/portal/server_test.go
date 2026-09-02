package portal

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentsv1 "github.com/chanzuckerberg/aws-oidc/api/v1"
	"github.com/chanzuckerberg/aws-oidc/pkg/awsaccess"
	"github.com/chanzuckerberg/aws-oidc/pkg/identity"
)

func TestTemplatesRender(t *testing.T) {
	s, err := NewServer(Config{})
	require.NoError(t, err)

	ent := &Entitlements{
		Accounts: []awsaccess.Account{{
			ID:    "111",
			Alias: "prod",
			Roles: []awsaccess.Role{{RoleARN: "arn:aws:iam::111:role/x", RoleName: "x"}},
		}},
		allowed: map[string]agentsv1.AWSGrant{},
	}

	// AWS page renders, and the current grant is pre-checked.
	rec := httptest.NewRecorder()
	s.render(rec, "agent_aws", pageData{
		Title:        "Edit",
		User:         &identity.User{Sub: "s", Email: "a@example.com"},
		Agent:        &agentsv1.Agent{ObjectMeta: metav1.ObjectMeta{Name: "bot"}},
		Entitlements: ent,
		Checked:      map[string]bool{"111|arn:aws:iam::111:role/x": true},
		Action:       "/agents/bot",
	})
	require.Equal(t, 200, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `type="checkbox"`)
	require.Contains(t, body, "checked")
	require.Contains(t, body, "prod")

	// List renders with an agent row.
	rec = httptest.NewRecorder()
	s.render(rec, "list", pageData{
		Title: "Your agents",
		User:  &identity.User{Sub: "s"},
		Agents: []agentsv1.Agent{{
			ObjectMeta: metav1.ObjectMeta{Name: "bot"},
			Spec: agentsv1.AgentSpec{
				Grants: []agentsv1.Grant{{AWS: &agentsv1.AWSGrant{AccountAlias: "prod", RoleName: "x"}}},
			},
		}},
	})
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "bot")
}
