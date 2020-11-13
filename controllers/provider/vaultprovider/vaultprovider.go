package vaultprovider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/pkg/errors"
)

var Type featuresv1alpha1.StackValueSourceType = featuresv1alpha1.StackValueSourceVault

type VaultProvider struct {
	config *featuresv1alpha1.StackValueSource
	path   string
}

func combineUrl(parts ...string) string {
	//take url parts with or without leading or trailing forward slash
	var urlParts []string
	for _, i := range parts {
		i = strings.Trim(i, "/")
		urlParts = append(urlParts, i)
	}
	return strings.Join(urlParts, "/")
}

func (p VaultProvider) Values() (interface{}, error) {
	url := combineUrl(p.config.Route, p.path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", string(p.config.Token))
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "call to vault failed")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Parse http response body")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vault returned status %s with the following body %s", resp.Status, body)
	}
	defer resp.Body.Close()

	// due to weirdness of vault response
	var tmp struct {
		Data struct {
			Data map[string]string
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &tmp); err != nil {
		return nil, errors.Wrap(err, "failed to marshal http response body")
	}

	// encoding returned value from StackValue
	for key, val := range tmp.Data.Data {
		tmp.Data.Data[key] = base64.StdEncoding.EncodeToString([]byte(val))
	}

	return tmp.Data.Data, nil
}

func New(c *featuresv1alpha1.StackValueSource, path string) VaultProvider {
	return VaultProvider{
		config: c,
		path:   path,
	}
}
