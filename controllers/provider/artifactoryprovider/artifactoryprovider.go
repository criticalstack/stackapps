package artifactoryprovider

import (
	"io/ioutil"
	"net/http"
	"strings"

	featuresv1alpha1 "github.com/criticalstack/stackapps/api/v1alpha1"
	"github.com/pkg/errors"
)

var Type featuresv1alpha1.StackValueSourceType = featuresv1alpha1.StackValueSourceArtifactory

type ArtifactoryProvider struct {
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

func (p ArtifactoryProvider) Values() (interface{}, error) {
	if len(p.config.Token) == 0 {
		return nil, errors.New("Please add Artifactory token to Token Secret")
	}
	url := combineUrl(p.config.Route, p.path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	token := string(p.config.Token)
	req.SetBasicAuth("admin", token) //make user configurable
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "call to artifactory failed")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Parse http artifactory response body")
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("artifactory returned status %q with the following body %q", resp.Status, body)
	}
	defer resp.Body.Close()

	return body, nil
}

func New(c *featuresv1alpha1.StackValueSource, path string) ArtifactoryProvider {
	return ArtifactoryProvider{
		config: c,
		path:   path,
	}
}
