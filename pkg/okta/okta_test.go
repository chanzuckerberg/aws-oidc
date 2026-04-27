package okta

import (
	"context"
	"testing"

	"github.com/okta/okta-sdk-golang/v3/okta"
	"github.com/stretchr/testify/require"
)

// oktaApplicationsWithNextField simulates a paginated Okta response.
// On the first call it returns pages[0] with a non-empty nextAfter cursor;
// on the second call it returns pages[1] with an empty cursor.
type oktaApplicationsWithNextField struct {
	pages     [][]okta.ListApplications200ResponseInner
	nextAfter []string // nextAfter[i] is the cursor returned after pages[i]
	index     int
}

func (m *oktaApplicationsWithNextField) ListApplications(_ context.Context, _, _ string) ([]okta.ListApplications200ResponseInner, string, error) {
	apps := m.pages[m.index]
	after := m.nextAfter[m.index]
	m.index++
	return apps, after, nil
}

type oktaApplicationsWithNilResponse struct{}

func (m *oktaApplicationsWithNilResponse) ListApplications(_ context.Context, _, _ string) ([]okta.ListApplications200ResponseInner, string, error) {
	apps := []okta.ListApplications200ResponseInner{
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id1")),
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id2")),
	}
	return apps, "", nil
}

// newOIDCApp is a helper that creates an OpenIdConnectApplication with the given ID.
func newOIDCApp(id string) *okta.OpenIdConnectApplication {
	app := okta.NewOpenIdConnectApplication()
	app.Id = &id
	return app
}

func TestGetClientIDs(t *testing.T) {
	r := require.New(t)
	apps := []okta.ListApplications200ResponseInner{
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id1")),
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id2")),
	}
	clientIDs, err := getClientIDsfromApplications(apps)
	r.NoError(err)
	r.Equal([]ClientID{"id1", "id2"}, clientIDs)
}

// TestMalformedOktaApps verifies that apps without an ID are excluded.
func TestMalformedOktaApps(t *testing.T) {
	r := require.New(t)
	apps := []okta.ListApplications200ResponseInner{
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(okta.NewOpenIdConnectApplication()),
		okta.OpenIdConnectApplicationAsListApplications200ResponseInner(okta.NewOpenIdConnectApplication()),
	}
	clientIDs, err := getClientIDsfromApplications(apps)
	r.NoError(err)
	r.Empty(clientIDs)
}

func TestPaginateListApplications(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	appInterfaces, err := paginateListApplications(ctx, "okta user id", &oktaApplicationsWithNilResponse{})
	r.NoError(err)
	r.Len(appInterfaces, 2)
}

func TestPaginateWithNext(t *testing.T) {
	ctx := context.Background()
	r := require.New(t)

	mock := &oktaApplicationsWithNextField{
		pages: [][]okta.ListApplications200ResponseInner{
			{
				okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id1")),
				okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id2")),
			},
			{
				okta.OpenIdConnectApplicationAsListApplications200ResponseInner(newOIDCApp("id3")),
			},
		},
		nextAfter: []string{"afterToken", ""},
	}

	clientIDs, err := GetClientIDs(ctx, "oktaUserID", mock)
	r.NoError(err)
	r.Equal([]ClientID{"id1", "id2", "id3"}, clientIDs)
}
