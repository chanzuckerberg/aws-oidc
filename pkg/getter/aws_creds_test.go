package getter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeSessionName(t *testing.T) {
	cases := map[string]struct {
		in   string
		want string
	}{
		"alphanumeric passthrough":       {"abc123XYZ", "abc123XYZ"},
		"valid punctuation passthrough":  {"a.b@c=d,e_f-g", "a.b@c=d,e_f-g"},
		"spaces become hyphens":          {"hello world", "hello-world"},
		"slash becomes hyphen":           {"name/with/slash", "name-with-slash"},
		"colon becomes hyphen":           {"role:session:1", "role-session-1"},
		"leading and trailing junk strip": {"!!!ok!!!", "ok"},
		"empty input":                    {"", "session"},
		"only invalid chars":             {"!!!", "session"},
		"client_id form":                 {"0oa1b2c3d4e5f6", "0oa1b2c3d4e5f6"},
		"long input truncated":           {strings.Repeat("a", 80), strings.Repeat("a", 64)},
		"email-style passthrough":        {"alice@example.com", "alice@example.com"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)
			got := sanitizeSessionName(tc.in)
			r.Equal(tc.want, got)
			r.LessOrEqual(len(got), stsSessionNameMaxLen)
			r.NotEmpty(got)
		})
	}
}
