package aws_config_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/chanzuckerberg/go-misc/oidc_cli/oidc_impl/client"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/propagation"
	"github.com/pkg/errors"
)

func RequestConfig(
	ctx context.Context,
	token *client.Token,
	configServiceURI string,
) (*server.AWSConfig, error) {
	ctx, span := beeline.StartSpan(ctx, "get_aws_config")
	defer span.Send()

	ver, err := util.VersionString()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, configServiceURI, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create request for %s", configServiceURI)
	}

	req.Header.Add(
		"Authorization",
		fmt.Sprintf("BEARER %s", token.IDToken),
	)
	req.Header.Set(
		"User-Agent",
		fmt.Sprintf("aws-oidc/%s", ver),
	)

	span.SerializeHeaders()
	req.Header.Add(
		propagation.TracePropagationHTTPHeader,
		span.SerializeHeaders(),
	)

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error requesting from %s", configServiceURI)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("non %d http status code received: %d", http.StatusOK, rsp.StatusCode)
	}
	defer rsp.Body.Close()

	body := bytes.NewBuffer(nil)
	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read request body: %w", err)
	}

	slog.Debug(fmt.Sprintf("request body: %s", body.String()))

	config := &server.AWSConfig{}
	err = json.Unmarshal(body.Bytes(), config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not json parse config from %s", configServiceURI)
	}

	return config, nil
}
