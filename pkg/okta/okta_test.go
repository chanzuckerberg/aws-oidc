package okta

import (
	"context"
	"net/http"
	"testing"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/stretchr/testify/require"
)

type oktaApplicationsWithNextField struct {
	ApplicationList      [][]okta.App
	ApplicationResponses []*okta.Response
	Index                int
}

func (oktaApp *oktaApplicationsWithNextField) ListApplications(ctx context.Context, qp *query.Params) ([]okta.App, *okta.Response, error) {
	appList := oktaApp.ApplicationList[oktaApp.Index] // get the next app off oktaApp.ApplicationList
	resp := oktaApp.ApplicationResponses[oktaApp.Index]

	oktaApp.Index = oktaApp.Index + 1
	return appList, resp, nil
}

type oktaApplicationsWithNilResponse struct{}

func (oktaApp *oktaApplicationsWithNilResponse) ListApplications(ctx context.Context, qp *query.Params) ([]okta.App, *okta.Response, error) {
	appList := []okta.App{
		&okta.Application{
			Id: "id1",
		},
		&okta.Application{
			Id: "id2",
		},
	}
	resp := &okta.Response{}

	return appList, resp, nil
}

func TestGetClientIDs(t *testing.T) {
	r := require.New(t)
	appInterfaces := []okta.App{
		&okta.Application{
			Id: "id1",
		},
		&okta.Application{
			Id: "id2",
		},
	}
	clientIDs, err := getClientIDsfromApplications(appInterfaces)
	r.NoError(err)
	r.Equal(clientIDs, []ClientID{"id1", "id2"})
}

// We're making an effort to exclude apps without a clientID
func TestMalformedOktaApps(t *testing.T) {
	r := require.New(t)

	appInterfacesNoId := []okta.App{
		&okta.Application{
			Label: "label1",
		},
		&okta.Application{
			Label: "label2",
		},
	}
	clientIDs, err := getClientIDsfromApplications(appInterfacesNoId)
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

	oktaApps := &oktaApplicationsWithNextField{}
	oktaApps.ApplicationList = append(
		oktaApps.ApplicationList,
		[]okta.App{
			&okta.Application{
				Id: "id1",
			},
			&okta.Application{
				Id: "id2",
			},
		},
		[]okta.App{
			&okta.Application{
				Id: "id3",
			},
		},
	)

	oktaApps.ApplicationResponses = append(
		oktaApps.ApplicationResponses,
		&okta.Response{
			Response: &http.Response{
				Header: http.Header{
					"Link": []string{"<localhost?after=afterToken&limit=2>; rel=\"next\""},
				},
			},
		},
		&okta.Response{
			Response: &http.Response{
				Header: http.Header{},
			},
		},
	)
	clientIDs, err := GetClientIDs(ctx, "oktaUserID", oktaApps)
	r.NoError(err)
	r.Equal(clientIDs, []ClientID{"id1", "id2", "id3"})
}
