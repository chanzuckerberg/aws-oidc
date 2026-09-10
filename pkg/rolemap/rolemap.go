// Package rolemap generates the aws-oidc rolemap — the mapping from Okta OIDC
// client IDs to assumable AWS role ARNs — by reading per-account Terraform state.
package rolemap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"gopkg.in/yaml.v2"
)

const (
	_accountIDOutputName             = "current_account_id"
	_accountAliasOutputName          = "current_account_alias"
	_oktaCZIAdminClientIDsOutputName = "czi-okta_okta_czi_admin_oidc_client_ids"
	_poweruserClientIDsOutputName    = "czi_okta_poweruser_oidc_client_ids"
	_readonlyClientIDsOutputName     = "czi_okta_readonly_oidc_client_ids"
	// _customRolesOutputName is a list of { role_name, client_ids } objects, one per custom
	// role in the account. Unlike the fixed poweruser/readonly/admin roles, the role name
	// varies, so the output carries it alongside the client IDs. Accounts with no custom
	// roles omit this output entirely, so a missing output yields no mappings.
	_customRolesOutputName = "czi_okta_custom_role_oidc_client_ids"

	_poweruserRoleARNFmt    = "arn:aws:iam::%s:role/poweruser"
	_readonlyRoleARNFmt     = "arn:aws:iam::%s:role/readonly"
	_oktaCZIAdminRoleARNFmt = "arn:aws:iam::%s:role/okta-czi-admin"
	_customRoleARNFmt       = "arn:aws:iam::%s:role/%s"
)

type s3Client interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type terraformState struct {
	Outputs map[string]terraformOutput `json:"outputs"`
}

type terraformOutput struct {
	Value json.RawMessage `json:"value"`
}

// Generate reads every Terraform state object under prefix and returns the full set of
// role mappings in deterministic order.
func Generate(ctx context.Context, client s3Client, bucket, prefix string) (okta.OIDCRoleMappings, error) {
	keys, err := listStateKeys(ctx, client, bucket, prefix)
	if err != nil {
		return nil, err
	}

	allMappings := make(okta.OIDCRoleMappings, 0)
	for _, key := range keys {
		// EvolutionaryScale manages its own identity integration.
		if strings.TrimSuffix(strings.TrimPrefix(key, prefix), ".tfstate") == "es-prod" {
			slog.Info("skipping state", "key", key)
			continue
		}

		mappings, err := objectRoleMappings(ctx, client, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("getting role mappings from s3://%s/%s: %w", bucket, key, err)
		}

		slog.Info("terraform state", "key", key, "mappingsCount", len(mappings))
		allMappings = append(allMappings, mappings...)
	}

	// Account ID and client ID alone are not enough:
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

func listStateKeys(ctx context.Context, client s3Client, bucket, prefix string) ([]string, error) {
	keys := make([]string, 0)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	for {
		output, err := client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("listing state objects in s3://%s/%s: %w", bucket, prefix, err)
		}
		for _, object := range output.Contents {
			key := aws.ToString(object.Key)
			if strings.HasSuffix(key, ".tfstate") {
				keys = append(keys, key)
			}
		}
		if !aws.ToBool(output.IsTruncated) {
			break
		}
		input.ContinuationToken = output.NextContinuationToken
	}

	sort.Strings(keys)
	return keys, nil
}

func objectRoleMappings(ctx context.Context, client s3Client, bucket, key string) (okta.OIDCRoleMappings, error) {
	object, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("reading object: %w", err)
	}
	defer object.Body.Close()

	return stateRoleMappings(object.Body)
}

func stateRoleMappings(reader io.Reader) (okta.OIDCRoleMappings, error) {
	var state terraformState
	if err := json.NewDecoder(reader).Decode(&state); err != nil {
		return nil, fmt.Errorf("decoding Terraform state: %w", err)
	}

	roleOutputNames := []string{
		_oktaCZIAdminClientIDsOutputName,
		_poweruserClientIDsOutputName,
		_readonlyClientIDsOutputName,
		_customRolesOutputName,
	}
	hasRoleOutputs := false
	for _, name := range roleOutputNames {
		if _, ok := state.Outputs[name]; ok {
			hasRoleOutputs = true
			break
		}
	}
	if !hasRoleOutputs {
		return nil, nil
	}

	var accountID string
	var accountAlias string
	if err := unmarshalOutput(state.Outputs, _accountIDOutputName, &accountID); err != nil {
		return nil, err
	}
	if err := unmarshalOutput(state.Outputs, _accountAliasOutputName, &accountAlias); err != nil {
		return nil, err
	}

	mappings := make(okta.OIDCRoleMappings, 0)
	fixedOutputs := []struct {
		name       string
		roleARNFmt string
	}{
		{name: _oktaCZIAdminClientIDsOutputName, roleARNFmt: _oktaCZIAdminRoleARNFmt},
		{name: _poweruserClientIDsOutputName, roleARNFmt: _poweruserRoleARNFmt},
		{name: _readonlyClientIDsOutputName, roleARNFmt: _readonlyRoleARNFmt},
	}
	for _, fixedOutput := range fixedOutputs {
		var clientIDs []string
		if err := unmarshalOptionalOutput(state.Outputs, fixedOutput.name, &clientIDs); err != nil {
			return nil, err
		}
		for _, clientID := range clientIDs {
			mapping, err := roleMapping(accountID, accountAlias, clientID, fmt.Sprintf(fixedOutput.roleARNFmt, accountID))
			if err != nil {
				return nil, err
			}
			mappings = append(mappings, mapping)
		}
	}

	if customOutput, ok := state.Outputs[_customRolesOutputName]; ok {
		var customRoles interface{}
		if err := json.Unmarshal(customOutput.Value, &customRoles); err != nil {
			return nil, fmt.Errorf("decoding %q output: %w", _customRolesOutputName, err)
		}
		customMappings, err := customRoleMappings(accountID, accountAlias, customRoles)
		if err != nil {
			return nil, fmt.Errorf("parsing custom role mappings: %w", err)
		}
		mappings = append(mappings, customMappings...)
	}

	return mappings, nil
}

func unmarshalOutput(outputs map[string]terraformOutput, name string, target interface{}) error {
	output, ok := outputs[name]
	if !ok {
		return fmt.Errorf("Terraform state has no %q output", name)
	}
	if err := json.Unmarshal(output.Value, target); err != nil {
		return fmt.Errorf("decoding %q output: %w", name, err)
	}
	return nil
}

func unmarshalOptionalOutput(outputs map[string]terraformOutput, name string, target interface{}) error {
	if _, ok := outputs[name]; !ok {
		return nil
	}
	return unmarshalOutput(outputs, name, target)
}

func roleMapping(accountID, accountAlias, clientID, roleARNValue string) (okta.OIDCRoleMapping, error) {
	roleARN, err := arn.Parse(roleARNValue)
	if err != nil {
		return okta.OIDCRoleMapping{}, fmt.Errorf("parsing role ARN: %w", err)
	}
	return okta.OIDCRoleMapping{
		OktaClientID:    clientID,
		AWSAccountID:    accountID,
		AWSAccountAlias: accountAlias,
		AWSRoleARN:      roleARN.String(),
	}, nil
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
		roleARN, err := arn.Parse(fmt.Sprintf(_customRoleARNFmt, accountID, roleName))
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
