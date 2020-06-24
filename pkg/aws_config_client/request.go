package aws_config_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	server "github.com/chanzuckerberg/aws-oidc/pkg/aws_config_server"
	"github.com/chanzuckerberg/go-misc/oidc_cli/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	server.AddBeelineFields(ctx, server.BeelineField{
		Key:   "aws_config_client http response status",
		Value: rsp.Status,
	})
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
		return nil, errors.Wrap(err, "could not read request body")
	}

	logrus.Debugf("request body: %s", body.String())

	config := &server.AWSConfig{}
	err = json.Unmarshal(body.Bytes(), config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not json parse config from %s", configServiceURI)
	}

	return config, nil
}
