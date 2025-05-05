package okta

import (
	"context"
	"fmt"
	"net/url"

	"github.com/honeycombio/beeline-go"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/peterhellberg/link"
)

// Configuration for an okta client
type OktaClientConfig struct {
	ClientID      string
	PrivateKeyPEM string
	OrgURL        string
}

func NewOktaClient(ctx context.Context, conf *OktaClientConfig) (*okta.Client, error) {
	_, client, err := okta.NewClient(
		ctx,
		okta.WithAuthorizationMode("PrivateKey"),
		okta.WithClientId(conf.ClientID),
		okta.WithScopes(([]string{"okta.apps.read"})),
		okta.WithPrivateKey(conf.PrivateKeyPEM),
		okta.WithOrgUrl(conf.OrgURL),
		okta.WithCache(true),
	)

	if err != nil {
		return nil, fmt.Errorf("error creating Okta client: %w", err)
	}
	return client, nil
}

func GetClientIDs(ctx context.Context, userID string, oktaClient AppResource) ([]ClientID, error) {
	ctx, span := beeline.StartSpan(ctx, "okta_get_client_ids")
	defer span.Send()

	apps, err := paginateListApplications(ctx, userID, oktaClient)
	if err != nil {
		return nil, err
	}
	return getClientIDsfromApplications(apps)
}

type AppResource interface {
	ListApplications(
		ctx context.Context,
		qp *query.Params,
	) ([]okta.App, *okta.Response, error)
}

func paginateListApplications(ctx context.Context, userID string, client AppResource) ([]okta.App, error) {
	ctx, span := beeline.StartSpan(ctx, "okta_list_applications")
	defer span.Send()

	var apps []okta.App

	qp := query.Params{
		Filter: fmt.Sprintf("user.id eq \"%s\"", userID),
	}

	for {
		currentApps, resp, err := client.ListApplications(ctx, &qp)
		if err != nil {
			return nil, fmt.Errorf("error listing applications: %w", err)
		}
		apps = append(apps, currentApps...)

		links := link.ParseResponse(resp.Response)

		if links["next"] == nil {
			return apps, nil // we're done, no next page
		}
		nextLink := links["next"].String()
		nextLinkURL, err := url.Parse(nextLink)
		if err != nil {
			return nil, fmt.Errorf("error parsing Link Header next url: %w", err)
		}

		nextLinkMapping := nextLinkURL.Query()
		qp.After = nextLinkMapping.Get("after")
	}
}

func getClientIDsfromApplications(appInterfaces []okta.App) ([]ClientID, error) {
	clientIDs := []ClientID{}
	for _, appInterface := range appInterfaces {
		// HACK(el): applications returned as interface which is useless...
		// 		type assertion back to concrete okta.Application
		app, ok := appInterface.(*okta.Application)
		if !ok {
			return nil, fmt.Errorf("appInterface not an Application")
		}
		if app.Id != "" {
			clientIDs = append(clientIDs, ClientID(app.Id))
		}
	}
	return clientIDs, nil
}
