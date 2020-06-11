package aws_config_client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/pkg/errors"
)

func RequestConfig(
	ctx context.Context,
	token *client.Token,
	configServiceURI string,
) (*server.AWSConfig, error) {
	req, err := http.NewRequest(http.MethodGet, configServiceURI, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create request for %s", configServiceURI)
	}

	req.Header.Add(
		"Authorization",
		fmt.Sprintf("BEARER %s", token.IDToken),
	)

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting from %s", configServiceURI)
	}
	defer rsp.Body.Close()

	config := &server.AWSConfig{}
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not json parse config from %s", configServiceURI)
	}

	return config, nil
}
