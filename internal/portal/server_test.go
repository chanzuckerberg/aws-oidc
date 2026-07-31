package portal

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplatesRender(t *testing.T) {
	s, err := NewServer(Config{})
	require.NoError(t, err)

	ent := &Entitlements{
		Accounts: []AccountAccess{{
			AccountID:    "111",
			AccountAlias: "prod",
			Roles:        []RoleAccess{{RoleARN: "arn:aws:iam::111:role/x", RoleName: "x"}},
		}},
		allowed: map[string]Grant{},
	}

	// Form renders, and the current grant is pre-checked.
	rec := httptest.NewRecorder()
	s.render(rec, "form", pageData{
		Title:        "Edit",
		User:         &User{Sub: "s", Email: "a@example.com"},
		Agent:        &Agent{Name: "bot"},
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
		Title:  "Your agents",
		User:   &User{Sub: "s"},
		Agents: []Agent{{Name: "bot", Grants: []Grant{{AccountAlias: "prod", RoleName: "x"}}}},
	})
	require.Equal(t, 200, rec.Code)
	require.Contains(t, rec.Body.String(), "bot")
}
