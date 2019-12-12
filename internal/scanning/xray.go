package scanning

import (
	"fmt"

	"github.com/target/go-arty/xray"
)

// XrayConfig contains all the information about the xray
type XrayConfig struct {
	Name     string `koanf:"name"`
	URL      string `koanf:"url"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Prefix   string `koanf:"prefix"`
}

// GetVulnerabilities gets vulnerabilities from xray
func (x XrayConfig) GetVulnerabilities(name, version string) ([]xray.SummaryArtifact, error) {
	url := "https://" + x.URL
	client, _ := xray.NewClient(url, nil)

	path := fmt.Sprintf("%s/%s/%s", x.Prefix, name, version)
	arty := &xray.SummaryArtifactRequest{
		Paths: &[]string{path},
	}
	client.Authentication.SetBasicAuth(x.Username, x.Password)
	sum, res, err := client.Summary.Artifact(arty)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Response code wrong [%v]", res.StatusCode)
	}
	if len(sum.GetErrors()) >= 1 {
		return nil, fmt.Errorf("Got an error from xray for [%s]m error [%s]", name, *sum.GetErrors()[0].Error)
	}
	return sum.GetArtifacts(), nil
}
