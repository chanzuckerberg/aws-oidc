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
)

func RequestConfig(
	ctx context.Context,
	token *client.Token,
	configServiceURI string,
) (*server.AWSConfig, error) {
	ver, err := util.VersionString()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, configServiceURI, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request for %s: %w", configServiceURI, err)
	}

	req.Header.Add(
		"Authorization",
		fmt.Sprintf("BEARER %s", token.IDToken),
	)
	req.Header.Set(
		"User-Agent",
		fmt.Sprintf("aws-oidc/%s", ver),
	)

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting from %s: %w", configServiceURI, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non %d http status code received: %d", http.StatusOK, rsp.StatusCode)
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
		return nil, fmt.Errorf("could not json parse config from %s: %w", configServiceURI, err)
	}

	return config, nil
}
