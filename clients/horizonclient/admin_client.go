package horizonclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/support/errors"
)

func (c *AdminClient) sendGetRequest(requestURL string, a interface{}) error {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return errors.Wrap(err, "error creating Admin HTTP request")
	}
	return c.sendHTTPRequest(req, a)
}

func (c *AdminClient) sendHTTPRequest(req *http.Request, a interface{}) error {
	if c.HTTP == nil {
		c.HTTP = http.DefaultClient
	}

	if c.horizonTimeout == 0 {
		c.horizonTimeout = HorizonTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.horizonTimeout)
	defer cancel()

	if resp, err := c.HTTP.Do(req.WithContext(ctx)); err != nil {
		return err
	} else {
		return decodeResponse(resp, a, req.URL.String(), c.clock)
	}
}

func (c *AdminClient) getIngestionFiltersURL(name string) (string, error) {
	baseURL, err := url.Parse("http://localhost")
	if err != nil {
		return "", err
	}
	baseURL.Path = baseURL.Path + "ingestion/filters/" + name
	adminPort := uint16(4200)
	if c.AdminPort > 0 {
		adminPort = c.AdminPort
	}
	baseURL.Host = fmt.Sprintf("%s:%d", baseURL.Hostname(), adminPort)
	return baseURL.String(), nil
}

func (c *AdminClient) AdminGetIngestionAssetFilter() (hProtocol.AssetFilterConfig, error) {
	url, err := c.getIngestionFiltersURL("asset")
	if err != nil {
		return hProtocol.AssetFilterConfig{}, err
	}
	var filter hProtocol.AssetFilterConfig
	err = c.sendGetRequest(url, &filter)
	return filter, err
}

func (c *AdminClient) AdminGetIngestionAccountFilter() (hProtocol.AccountFilterConfig, error) {
	url, err := c.getIngestionFiltersURL("account")
	if err != nil {
		return hProtocol.AccountFilterConfig{}, err
	}
	var filter hProtocol.AccountFilterConfig
	err = c.sendGetRequest(url, &filter)
	return filter, err
}

func (c *AdminClient) AdminSetIngestionAssetFilter(filter hProtocol.AssetFilterConfig) error {
	url, err := c.getIngestionFiltersURL("asset")
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(filter)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, buf)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	req.Header.Add("Content-Type", "application/json")
	return c.sendHTTPRequest(req, nil)
}

func (c *AdminClient) AdminSetIngestionAccountFilter(filter hProtocol.AccountFilterConfig) error {
	url, err := c.getIngestionFiltersURL("account")
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(filter)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, buf)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	req.Header.Add("Content-Type", "application/json")
	return c.sendHTTPRequest(req, nil)
}

// ensure that the horizon admin client implements AdminClientInterface
var _ AdminClientInterface = &AdminClient{}
