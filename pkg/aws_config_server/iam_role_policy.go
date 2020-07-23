package aws_config_server

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
	"github.com/pkg/errors"
)

const assumeRoleWebIdentityAction = "sts:AssumeRoleWithWebIdentity"

type PolicyDocument struct {
	Version    string           `json:"Version"`
	Statements []StatementEntry `json:"Statement"`
}

type StatementEntry struct {
	Effect    string    `json:"Effect"`
	Action    Action    `json:"Action"`
	Sid       string    `json:"Sid"`
	Principal Principal `json:"Principal"`
	Condition Condition `json:"Condition"`
}

func (se *StatementEntry) GetFederatedClientIDs(oidcProviderHostname string) []okta.ClientID {
	if se == nil {
		return nil
	}

	clientKey := fmt.Sprintf("%s:aud", oidcProviderHostname)
	clientIDs, ok := se.Condition.StringEquals[clientKey]
	if !ok {
		return nil
	}

	isWebIdentityAction := false
	for _, action := range se.Action {
		isWebIdentityAction = isWebIdentityAction || (action == assumeRoleWebIdentityAction)
	}
	if !isWebIdentityAction {
		return nil
	}

	filteredClientIDs := []okta.ClientID{}
	// Loop through all the clientIDS for thsi statement's conditional
	for _, clientIDStr := range clientIDs {
		clientID := okta.ClientID(clientIDStr)

		filteredClientIDs = append(filteredClientIDs, clientID)
	}
	return filteredClientIDs
}

// We only care about the "StringEquals" field in Condition
type Condition struct {
	StringEquals StringEqualsCondition `json:"StringEquals"`
}

type StringEqualsCondition map[string][]string

func (sec *StringEqualsCondition) UnmarshalJSON(data []byte) error {
	untypedMap := map[string]interface{}{}
	err := json.Unmarshal(data, &untypedMap)
	if err != nil {
		return err
	}

	returnMap := map[string][]string{}

	for key, val := range untypedMap {
		str, ok := val.(string)
		if ok {
			returnMap[key] = []string{str}
			continue
		}

		strSlice := []string{}

		slice, ok := val.([]interface{})
		if !ok {
			return errors.Errorf("unrecognized type %v", val)
		}

		for _, maybeStr := range slice {
			str, ok := maybeStr.(string)
			if !ok {
				return errors.Errorf("unrecognized type %v", val)
			}

			strSlice = append(strSlice, str)
		}

		returnMap[key] = strSlice
	}

	*sec = returnMap
	return nil
}

// We only care about the "Federated" field in Principal
type Principal struct {
	Federated string `json:"Federated"`
}

type Action []string

func (a *Action) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		*a = []string{str}
		return nil
	}
	// If the error is not an unmarshal type error, then we return the error
	if _, ok := err.(*json.UnmarshalTypeError); err != nil && !ok {
		return errors.Wrap(err, "Unexpected error type from unmarshaling")
	}

	var strSlice []string
	err = json.Unmarshal(data, &strSlice)
	if err == nil {
		*a = strSlice
		return nil
	}
	return errors.Wrap(err, "Unable to unmarshal Action")
}

func NewPolicyDocument(assumeRolePolicyDocument string) (*PolicyDocument, error) {
	// the IAM Role outputs a url-encoded policy document, so we need to escape characters
	policyStr, err := url.PathUnescape(assumeRolePolicyDocument)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to escape URL encoding")
	}
	policyDoc := &PolicyDocument{}
	err = json.Unmarshal([]byte(policyStr), policyDoc)
	return policyDoc, errors.Wrapf(err, "Unable to unmarshal policy document to struct policy: %s", policyStr)
}
