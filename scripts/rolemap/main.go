package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/go-tfe"
	"gopkg.in/yaml.v2"
)

type OIDCRoleMapping struct {
	AWSAccountID string `yaml:"aws_account_id"`
	AWSRoleARN   string `yaml:"aws_role_arn"`
	OktaClientID string `yaml:"okta_client_id"`
}

const (
	accountIDOutputName             = "current_account_id"
	oktaCZIAdminClientIDsOutputName = "okta-czi-admin_oidc_client_ids"
	poweruserClientIDsOutputName    = "poweruser_oidc_client_ids"
	readonlyClientIDsOutputName     = "readonly_oidc_client_ids"

	poweruserRoleARNFmt    = "arn:aws:iam::%s:role/poweruser"
	readonlyRoleARNFmt     = "arn:aws:iam::%s:role/readonly"
	oktaCZIAdminRoleARNFmt = "arn:aws:iam::%s:role/okta-czi-admin"
)

func workspaceRoleMappings(ctx context.Context, client *tfe.Client, workspaceID string) ([]OIDCRoleMapping, error) {
	currentState, err := client.StateVersionOutputs.ReadCurrent(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("reading current state version outputs: %w", err)
	}

	mappings := []OIDCRoleMapping{}
	var accountID string
	for _, output := range currentState.Items {
		switch output.Name {
		case accountIDOutputName:
			accountID = output.Value.(string)
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
				mappings = append(mappings, OIDCRoleMapping{
					OktaClientID: clientID.(string),
					AWSAccountID: accountID,
					AWSRoleARN:   roleARN.String(),
				})
			}
		case poweruserClientIDsOutputName:
			for _, clientID := range output.Value.([]interface{}) {
				roleARN, err := arn.Parse(fmt.Sprintf(poweruserRoleARNFmt, accountID))
				if err != nil {
					return nil, fmt.Errorf("parsing role ARN: %w", err)
				}
				mappings = append(mappings, OIDCRoleMapping{
					OktaClientID: clientID.(string),
					AWSAccountID: accountID,
					AWSRoleARN:   roleARN.String(),
				})
			}
		case readonlyClientIDsOutputName:
			for _, clientID := range output.Value.([]interface{}) {
				roleARN, err := arn.Parse(fmt.Sprintf(readonlyRoleARNFmt, accountID))
				if err != nil {
					return nil, fmt.Errorf("parsing role ARN: %w", err)
				}
				mappings = append(mappings, OIDCRoleMapping{
					OktaClientID: clientID.(string),
					AWSAccountID: accountID,
					AWSRoleARN:   roleARN.String(),
				})
			}
		}
	}

	return mappings, nil
}

func getAllAccountTFEWorkspaces(ctx context.Context, client *tfe.Client) ([]*tfe.Workspace, error) {
	workspaces := []*tfe.Workspace{}
	page := 1
	pageSize := 100
	for {
		workspaceListPage, err := client.Workspaces.List(ctx, "shared-infra", &tfe.WorkspaceListOptions{
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

func exec(ctx context.Context) error {
	token := os.Getenv("TFE_TOKEN")
	if token == "" {
		return fmt.Errorf("TFE_TOKEN environment variable must be set.")
	}

	config := &tfe.Config{
		Token:   token,
		Address: "https://si.prod.tfe.czi.technology",
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return fmt.Errorf("creating new TFE client: %w", err)
	}

	workspaces, err := getAllAccountTFEWorkspaces(ctx, client)
	if err != nil {
		return fmt.Errorf("getting all shared-infra account-* workspaces: %w", err)
	}

	allMappings := []OIDCRoleMapping{}
	for _, workspace := range workspaces {
		mappings, err := workspaceRoleMappings(ctx, client, workspace.ID)
		if err != nil {
			return fmt.Errorf("getting role mappings for workspace %s: %w", workspace.Name, err)
		}

		slog.Info("workspace", "name", workspace.Name, "id", workspace.ID, "mappingsCounts", len(mappings))
		allMappings = append(allMappings, mappings...)
	}

	b, err := yaml.Marshal(allMappings)
	if err != nil {
		return fmt.Errorf("marshalling role mappings to YAML: %w", err)
	}

	for _, env := range []string{"rdev", "prod"} {
		err = os.WriteFile(fmt.Sprintf("../../.infra/%s/rolemap/rolemap.yaml", env), b, 0644)
		if err != nil {
			return fmt.Errorf("writing role mappings to file: %w", err)
		}
	}

	return nil
}

func main() {
	err := exec(context.Background())
	if err != nil {
		panic(err)
	}
}
