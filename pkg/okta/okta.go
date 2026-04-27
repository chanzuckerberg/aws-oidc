package okta

import (
	"context"
	"fmt"
	"net/url"

	"github.com/okta/okta-sdk-golang/v3/okta"
)

// OktaClientConfig holds configuration for creating an Okta API client.
type OktaClientConfig struct {
	ClientID      string
	PrivateKeyPEM string
	OrgURL        string
}

// AppLister abstracts Okta application listing for testability.
// ListApplications returns the page of apps matching filter, the after-cursor
// for the next page (empty string when there is no next page), and any error.
type AppLister interface {
	ListApplications(ctx context.Context, filter, after string) ([]okta.ListApplications200ResponseInner, string, error)
}

// OktaAppClient wraps the v3 ApplicationAPI and implements AppLister.
type OktaAppClient struct {
	api okta.ApplicationAPI
}

var _ AppLister = (*OktaAppClient)(nil)

func (c *OktaAppClient) ListApplications(ctx context.Context, filter, after string) ([]okta.ListApplications200ResponseInner, string, error) {
	req := c.api.ListApplications(ctx).Filter(filter)
	if after != "" {
		req = req.After(after)
	}
	apps, resp, err := req.Execute()
	if err != nil {
		return nil, "", err
	}
	if !resp.HasNextPage() {
		return apps, "", nil
	}
	nextURL := resp.NextPage()
	parsedURL, err := url.Parse(nextURL)
	if err != nil {
		return nil, "", fmt.Errorf("parsing next page URL: %w", err)
	}
	return apps, parsedURL.Query().Get("after"), nil
}

// NewOktaClient creates an Okta API client and returns it as an AppResource.
// The v3 SDK supports both PKCS#1 (BEGIN RSA PRIVATE KEY) and
// PKCS#8 (BEGIN PRIVATE KEY) PEM formats.
func NewOktaClient(ctx context.Context, conf *OktaClientConfig) (AppLister, error) {
	cfg, err := okta.NewConfiguration(
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(conf.ClientID),
		okta.WithScopes([]string{"okta.apps.read"}),
		okta.WithPrivateKey(conf.PrivateKeyPEM),
		okta.WithOrgUrl(conf.OrgURL),
		okta.WithCache(true),
	)
	if err != nil {
		return nil, fmt.Errorf("creating Okta configuration: %w", err)
	}
	client := okta.NewAPIClient(cfg)
	return &OktaAppClient{api: client.ApplicationAPI}, nil
}

func GetClientIDs(ctx context.Context, userID string, oktaClient AppLister) ([]ClientID, error) {
	apps, err := paginateListApplications(ctx, userID, oktaClient)
	if err != nil {
		return nil, err
	}
	return getClientIDsfromApplications(apps)
}

func paginateListApplications(ctx context.Context, userID string, client AppLister) ([]okta.ListApplications200ResponseInner, error) {
	var apps []okta.ListApplications200ResponseInner
	filter := fmt.Sprintf("user.id eq \"%s\"", userID)
	after := ""

	for {
		currentApps, nextAfter, err := client.ListApplications(ctx, filter, after)
		if err != nil {
			return nil, fmt.Errorf("error listing applications: %w", err)
		}
		apps = append(apps, currentApps...)

		if nextAfter == "" {
			return apps, nil
		}
		after = nextAfter
	}
}

// appIDer is satisfied by all concrete Okta application types, which embed
// okta.Application and therefore inherit its GetId method.
type appIDer interface {
	GetId() string
}

func getClientIDsfromApplications(appInterfaces []okta.ListApplications200ResponseInner) ([]ClientID, error) {
	clientIDs := make([]ClientID, 0, len(appInterfaces))
	for _, appInterface := range appInterfaces {
		instance := appInterface.GetActualInstance()
		if instance == nil {
			continue
		}
		idProvider, ok := instance.(appIDer)
		if !ok {
			return nil, fmt.Errorf("unexpected application type %T", instance)
		}
		id := idProvider.GetId()
		if id != "" {
			clientIDs = append(clientIDs, ClientID(id))
		}
	}
	return clientIDs, nil
}
