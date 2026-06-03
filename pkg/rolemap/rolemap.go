// Package rolemap generates the aws-oidc rolemap — the mapping from Okta OIDC
// client IDs to assumable AWS role ARNs — by reading the per-account Terraform
// state outputs published to TFE.
package rolemap

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/hashicorp/go-tfe"
	"gopkg.in/yaml.v2"
)

const (
	tfeAddress = "https://si.prod.tfe.czi.technology"
	tfeOrg     = "shared-infra"

	accountIDOutputName             = "current_account_id"
	accountAliasOutputName          = "current_account_alias"
	oktaCZIAdminClientIDsOutputName = "czi-okta_okta_czi_admin_oidc_client_ids"
	poweruserClientIDsOutputName    = "czi_okta_poweruser_oidc_client_ids"
	readonlyClientIDsOutputName     = "czi_okta_readonly_oidc_client_ids"
	// customRolesOutputName is a list of { role_name, client_ids } objects, one per custom
	// role in the account. Unlike the fixed poweruser/readonly/admin roles, the role name
	// varies, so the output carries it alongside the client IDs. Accounts with no custom
	// roles omit this output entirely, so a missing output yields no mappings.
	customRolesOutputName = "czi_okta_custom_role_oidc_client_ids"

	poweruserRoleARNFmt    = "arn:aws:iam::%s:role/poweruser"
	readonlyRoleARNFmt     = "arn:aws:iam::%s:role/readonly"
	oktaCZIAdminRoleARNFmt = "arn:aws:iam::%s:role/okta-czi-admin"
	customRoleARNFmt       = "arn:aws:iam::%s:role/%s"
)

// Generate reads every shared-infra accounts-* workspace's current state from TFE and
// returns the full set of role mappings, in a deterministic order.
func Generate(ctx context.Context, tfeToken string) (okta.OIDCRoleMappings, error) {
	client, err := tfe.NewClient(&tfe.Config{Token: tfeToken, Address: tfeAddress})
	if err != nil {
		return nil, fmt.Errorf("creating new TFE client: %w", err)
	}

	workspaces, err := getAllAccountTFEWorkspaces(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("getting all %s account-* workspaces: %w", tfeOrg, err)
	}

	allMappings := okta.OIDCRoleMappings{}
	for _, workspace := range workspaces {
		// Skip specific workspaces
		if workspace.Name == "accounts-es-prod" {
			slog.Info("skipping workspace", "name", workspace.Name)
			continue
		}

		mappings, err := workspaceRoleMappings(ctx, client, workspace.ID)
		if err != nil {
			return nil, fmt.Errorf("getting role mappings for workspace %s: %w", workspace.Name, err)
		}

		slog.Info("workspace", "name", workspace.Name, "id", workspace.ID, "mappingsCounts", len(mappings))
		allMappings = append(allMappings, mappings...)
	}

	// Total ordering so the generated YAML is deterministic regardless of the order TFE
	// returns workspaces or state outputs. Account ID and client ID alone are not enough:
	// one client ID can map to more than one role ARN (e.g. a custom role plus poweruser),
	// so without the role-ARN tiebreaker those rows could swap places and churn the diff.
	sort.Slice(allMappings, func(i, j int) bool {
		a, b := allMappings[i], allMappings[j]
		if a.AWSAccountID != b.AWSAccountID {
			return a.AWSAccountID > b.AWSAccountID
		}
		if a.OktaClientID != b.OktaClientID {
			return a.OktaClientID > b.OktaClientID
		}
		return a.AWSRoleARN > b.AWSRoleARN
	})

	return allMappings, nil
}

// Marshal renders mappings to the YAML form stored in the rolemap ConfigMap.
func Marshal(mappings okta.OIDCRoleMappings) ([]byte, error) {
	b, err := yaml.Marshal(mappings)
	if err != nil {
		return nil, fmt.Errorf("marshalling role mappings to YAML: %w", err)
	}
	return b, nil
}

func workspaceRoleMappings(ctx context.Context, client *tfe.Client, workspaceID string) (okta.OIDCRoleMappings, error) {
	currentState, err := client.StateVersionOutputs.ReadCurrent(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("reading current state version outputs: %w", err)
	}

	mappings := okta.OIDCRoleMappings{}
	var accountID string
	var accountAlias string
	for _, output := range currentState.Items {
		switch output.Name {
		case accountIDOutputName:
			accountID = output.Value.(string)
		case accountAliasOutputName:
			accountAlias = output.Value.(string)
		}
	}

	for _, output := range currentState.Items {
		switch output.Name {
		case oktaCZIAdminClientIDsOutputName:
			for _, clientID := range output.Value.([]interface{}) {
				roleARN, err := arn.Parse(fmt.Sprintf(oktaCZIAdminRoleARNFmt, accountID))
				if err != nil {
					return nil, fmt.Errorf("parsing role ARN: %w", err)
				}
				mappings = append(mappings, okta.OIDCRoleMapping{
					OktaClientID:    clientID.(string),
					AWSAccountID:    accountID,
					AWSAccountAlias: accountAlias,
					AWSRoleARN:      roleARN.String(),
				})
			}
		case poweruserClientIDsOutputName:
			for _, clientID := range output.Value.([]interface{}) {
				roleARN, err := arn.Parse(fmt.Sprintf(poweruserRoleARNFmt, accountID))
				if err != nil {
					return nil, fmt.Errorf("parsing role ARN: %w", err)
				}
				mappings = append(mappings, okta.OIDCRoleMapping{
					OktaClientID:    clientID.(string),
					AWSAccountID:    accountID,
					AWSAccountAlias: accountAlias,
					AWSRoleARN:      roleARN.String(),
				})
			}
		case readonlyClientIDsOutputName:
			for _, clientID := range output.Value.([]interface{}) {
				roleARN, err := arn.Parse(fmt.Sprintf(readonlyRoleARNFmt, accountID))
				if err != nil {
					return nil, fmt.Errorf("parsing role ARN: %w", err)
				}
				mappings = append(mappings, okta.OIDCRoleMapping{
					OktaClientID:    clientID.(string),
					AWSAccountID:    accountID,
					AWSAccountAlias: accountAlias,
					AWSRoleARN:      roleARN.String(),
				})
			}
		case customRolesOutputName:
			customMappings, err := customRoleMappings(accountID, accountAlias, output.Value)
			if err != nil {
				return nil, fmt.Errorf("parsing custom role mappings: %w", err)
			}
			mappings = append(mappings, customMappings...)
		}
	}

	return mappings, nil
}

// customRoleMappings parses the czi_okta_custom_role_oidc_client_ids output, a list of
// { role_name, client_ids } objects, into role mappings. It tolerates anything that does
// not match that shape by skipping it, so an account without the output (or with an
// unexpected value) simply contributes no custom-role mappings rather than failing the run.
func customRoleMappings(accountID, accountAlias string, value interface{}) (okta.OIDCRoleMappings, error) {
	entries, ok := value.([]interface{})
	if !ok {
		return nil, nil
	}

	mappings := okta.OIDCRoleMappings{}
	for _, entry := range entries {
		obj, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		roleName, ok := obj["role_name"].(string)
		if !ok || roleName == "" {
			continue
		}
		clientIDs, ok := obj["client_ids"].([]interface{})
		if !ok {
			continue
		}
		roleARN, err := arn.Parse(fmt.Sprintf(customRoleARNFmt, accountID, roleName))
		if err != nil {
			return nil, fmt.Errorf("parsing role ARN for %q: %w", roleName, err)
		}
		for _, clientID := range clientIDs {
			id, ok := clientID.(string)
			if !ok {
				continue
			}
			mappings = append(mappings, okta.OIDCRoleMapping{
				OktaClientID:    id,
				AWSAccountID:    accountID,
				AWSAccountAlias: accountAlias,
				AWSRoleARN:      roleARN.String(),
			})
		}
	}

	return mappings, nil
}

func getAllAccountTFEWorkspaces(ctx context.Context, client *tfe.Client) ([]*tfe.Workspace, error) {
	workspaces := []*tfe.Workspace{}
	page := 1
	pageSize := 100
	for {
		workspaceListPage, err := client.Workspaces.List(ctx, tfeOrg, &tfe.WorkspaceListOptions{
			Search: "accounts-",
			ListOptions: tfe.ListOptions{
				PageNumber: page,
				PageSize:   pageSize,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("listing workspaces: %w", err)
		}
		workspaces = append(workspaces, workspaceListPage.Items...)
		if len(workspaceListPage.Items) < pageSize {
			break
		}
		page++
	}

	return workspaces, nil
}
