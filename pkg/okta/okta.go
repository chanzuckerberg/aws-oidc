package okta

import (
	"context"
	"net/url"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/peterhellberg/link"
	"github.com/pkg/errors"
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
		okta.WithCache(false),
	)

	return client, errors.Wrap(err, "error creating Okta client")
}

func GetClientIDs(ctx context.Context, oktaClient AppResource) ([]ClientID, error) {
	apps, err := paginateListApplications(ctx, oktaClient)
	if err != nil {
		return nil, err
	}
	return getClientIDsfromApplications(ctx, apps)
}

type AppResource interface {
	ListApplications(context.Context, *query.Params) ([]okta.App, *okta.Response, error)
}

func paginateListApplications(ctx context.Context, client AppResource) ([]okta.App, error) {
	var qp query.Params
	var apps []okta.App

	for {
		currentApps, resp, err := client.ListApplications(ctx, &qp)
		if err != nil {
			return nil, errors.Wrap(err, "error listing applications")
		}
		apps = append(apps, currentApps...)

		links := link.ParseResponse(resp.Response)

		if links["next"] == nil {
			return apps, nil // we're done, no next page
		}
		nextLink := links["next"].String()
		nextLinkURL, err := url.Parse(nextLink)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing Link Header next url")
		}

		nextLinkMapping := nextLinkURL.Query()
		qp.After = nextLinkMapping.Get("after")
	}
}

func getClientIDsfromApplications(
	ctx context.Context,
	appInterfaces []okta.App) ([]ClientID, error) {
	clientIDs := []ClientID{}
	for _, appInterface := range appInterfaces {
		// HACK(el): applications returned as interface which is useless...
		// 		type assertion back to concrete okta.Application
		app, ok := appInterface.(*okta.Application)
		if !ok {
			return nil, errors.New("appInterface not an Application")
		}
		if app.Id != "" {
			clientIDs = append(clientIDs, ClientID(app.Id))
		}
	}
	return clientIDs, nil
}
