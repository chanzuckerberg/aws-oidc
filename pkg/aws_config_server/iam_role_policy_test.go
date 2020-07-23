package aws_config_server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/stretchr/testify/require"
)

func TestConditionUnmarshalArray(t *testing.T) {
	r := require.New(t)

	conditionExpected := &Condition{
		StringEquals: StringEqualsCondition(
			map[string][]string{
				"foo:aud": {"1", "2", "3"},
			},
		),
	}
	conditionStr := `{"StringEquals":{"foo:aud":["1","2","3"]}}`

	condition := &Condition{}
	err := json.Unmarshal([]byte(conditionStr), condition)
	r.NoError(err)

	r.Equal(conditionExpected, condition)
}

func TestConditionUnmarshalString(t *testing.T) {
	r := require.New(t)

	conditionExpected := &Condition{
		StringEquals: StringEqualsCondition(
			map[string][]string{
				"foo:aud": {"1"},
			},
		),
	}
	conditionStr := `{"StringEquals":{"foo:aud":"1"}}`

	condition := &Condition{}
	err := json.Unmarshal([]byte(conditionStr), condition)
	r.NoError(err)

	r.Equal(conditionExpected, condition)
}

func TestActionUnmarshalArray(t *testing.T) {
	r := require.New(t)

	actionExpected := Action([]string{"1", "2", "3"})

	actionStr := `["1","2","3"]`

	action := Action{}
	err := json.Unmarshal([]byte(actionStr), &action)
	r.NoError(err)

	r.Equal(actionExpected, action)
}

func TestActionUnmarshalString(t *testing.T) {
	r := require.New(t)

	actionExpected := Action([]string{"1"})

	actionStr := `"1"`

	action := Action{}
	err := json.Unmarshal([]byte(actionStr), &action)
	r.NoError(err)

	r.Equal(actionExpected, action)
}

func TestStatementGetFederatedClientIDs(t *testing.T) {
	r := require.New(t)

	type test struct {
		name     string
		se       *StatementEntry
		expected []okta.ClientID
	}

	oidcHostname := "testing-hostname"

	tests := []test{
		{
			name:     "nil se",
			se:       nil,
			expected: nil,
		},
		{
			name: "audience not match",
			se: &StatementEntry{
				Condition: Condition{
					StringEquals: StringEqualsCondition(map[string][]string{
						"no-match:aud": {"1", "2", "3"},
					}),
				},
			},
			expected: nil,
		},
		{
			name: "audience match, not web identity",
			se: &StatementEntry{
				Condition: Condition{
					StringEquals: StringEqualsCondition(map[string][]string{
						"testing-hostname:aud": {"1", "2", "3"},
					}),
				},
			},
			expected: nil,
		},
		{
			name: "audience match, web identity",
			se: &StatementEntry{
				Action: Action{"don't care", assumeRoleWebIdentityAction},
				Condition: Condition{
					StringEquals: StringEqualsCondition(map[string][]string{
						"testing-hostname:aud": {"1", "2", "3"},
					}),
				},
			},
			expected: []okta.ClientID{"1", "2", "3"},
		},
	}

	for _, test := range tests {
		fmt.Println(test.name)
		r.Equal(test.expected, test.se.GetFederatedClientIDs(oidcHostname))
	}
}

func TestNewPolicyDocument(t *testing.T) {
	r := require.New(t)

	expectedPolicy := &PolicyDocument{
		Statements: []StatementEntry{
			{
				Action: Action{},
				Condition: Condition{
					StringEquals: StringEqualsCondition{},
				},
			},
		},
	}

	jsonPolicy, err := json.Marshal(expectedPolicy)
	r.NoError(err)

	escapedPolicy := url.PathEscape(string(jsonPolicy))

	policy, err := NewPolicyDocument(escapedPolicy)
	r.NoError(err)
	r.Equal(expectedPolicy, policy)
}
