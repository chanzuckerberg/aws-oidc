// Package catalog holds the curated set of grantable policies. An agent grant may only
// reference a catalog entry, never a free-form policy, so this is the closed set of access
// an agent role can ever receive. It is loaded from an editable config source (a file or a
// ConfigMap in the style of the rolemap) so the catalog changes without a rebuild.
package catalog

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Policy is one curated, grantable managed policy. The operator attaches exactly one of
// these to each per-agent role.
type Policy struct {
	// ID is the stable identifier a grant references (for example "s3-readonly").
	ID string `yaml:"id"`
	// DisplayName is shown in the portal dropdown.
	DisplayName string `yaml:"display_name"`
	// PolicyName is the IAM managed policy name. For a customer-managed policy it resolves
	// per account; for an AWS-managed policy set AWSManaged and it resolves under the aws
	// account alias.
	PolicyName string `yaml:"policy_name"`
	// AWSManaged selects the AWS-managed ARN namespace (arn:aws:iam::aws:policy/...).
	AWSManaged bool `yaml:"aws_managed,omitempty"`
}

// ARN returns the managed policy ARN for the given target account. AWS-managed policies
// ignore the account and resolve under the shared aws namespace.
func (p Policy) ARN(accountID string) string {
	if p.AWSManaged {
		return fmt.Sprintf("arn:aws:iam::aws:policy/%s", p.PolicyName)
	}
	return fmt.Sprintf("arn:aws:iam::%s:policy/%s", accountID, p.PolicyName)
}

// Catalog is the set of grantable policies keyed by ID.
type Catalog struct {
	policies map[string]Policy
}

// catalogFile is the on-disk shape.
type catalogFile struct {
	Policies []Policy `yaml:"policies"`
}

// Load parses a catalog from YAML. It rejects duplicate or empty IDs so a typo in the
// config cannot silently shadow a policy.
func Load(raw []byte) (*Catalog, error) {
	file := catalogFile{}
	err := yaml.Unmarshal(raw, &file)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling catalog: %w", err)
	}

	policies := make(map[string]Policy, len(file.Policies))
	for _, p := range file.Policies {
		if p.ID == "" {
			return nil, fmt.Errorf("catalog entry with empty id")
		}
		if p.PolicyName == "" {
			return nil, fmt.Errorf("catalog entry %q has empty policy_name", p.ID)
		}
		_, dup := policies[p.ID]
		if dup {
			return nil, fmt.Errorf("duplicate catalog id %q", p.ID)
		}
		policies[p.ID] = p
	}
	return &Catalog{policies: policies}, nil
}

// Get returns the policy for an id.
func (c *Catalog) Get(id string) (Policy, bool) {
	p, ok := c.policies[id]
	return p, ok
}

// Policies returns the catalog entries. The order is unspecified.
func (c *Catalog) Policies() []Policy {
	out := make([]Policy, 0, len(c.policies))
	for _, p := range c.policies {
		out = append(out, p)
	}
	return out
}
