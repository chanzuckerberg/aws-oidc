package aws_config_server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chanzuckerberg/aws-oidc/pkg/okta"
)

type AccessResolver struct {
	oidcProvider     string
	oktaAppClient    okta.AppLister
	mappingsProvider MappingsProvider
}

func NewAccessResolver(
	oidcProvider string,
	oktaAppClient okta.AppLister,
	mappingsProvider MappingsProvider,
) (*AccessResolver, error) {
	if strings.TrimSpace(oidcProvider) == "" {
		return nil, errors.New("OIDC provider is required")
	}
	if oktaAppClient == nil {
		return nil, errors.New("Okta application client is required")
	}
	if mappingsProvider == nil {
		return nil, errors.New("mappings provider is required")
	}
	return &AccessResolver{
		oidcProvider:     normalizeIssuer(oidcProvider),
		oktaAppClient:    oktaAppClient,
		mappingsProvider: mappingsProvider,
	}, nil
}

func (r *AccessResolver) Resolve(ctx context.Context, userID string) (*AWSConfig, error) {
	mappings, err := r.mappingsProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("read rolemap: %w", err)
	}
	clientIDs, err := okta.GetClientIDs(ctx, userID, r.oktaAppClient)
	if err != nil {
		return nil, fmt.Errorf("get assigned applications: %w", err)
	}
	config, err := createAWSConfig(r.oidcProvider, mappings, clientIDs)
	if err != nil {
		return nil, fmt.Errorf("create AWS config: %w", err)
	}
	return config, nil
}

func (r *AccessResolver) OIDCProvider() string {
	return r.oidcProvider
}
