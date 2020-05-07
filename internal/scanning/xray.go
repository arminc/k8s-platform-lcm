package scanning

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/target/go-arty/xray"
)

// XrayConfig contains all the information about the xray
type XrayConfig struct {
	Name     string   `koanf:"name"`
	URL      string   `koanf:"url"`
	Username string   `koanf:"username"`
	Password string   `koanf:"password"`
	Prefixes []Prefix `koanf:"prefixes"`
}

// Prefix information about the index used by Xray
type Prefix struct {
	Prefix string   `koanf:"prefix"`
	Images []string `koanf:"images"`
}

// GetVulnerabilities gets vulnerabilities from xray
func (x XrayConfig) GetVulnerabilities(name, version string) ([]xray.SummaryArtifact, error) {
	url := "https://" + x.URL
	client, _ := xray.NewClient(url, nil)

	path := fmt.Sprintf("%s/%s/%s", x.getPrefix(name), name, version)
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

func (x XrayConfig) getPrefix(name string) string {
	if len(x.Prefixes) == 1 {
		return x.Prefixes[0].Prefix
	}
	for _, prefix := range x.Prefixes {
		for _, image := range prefix.Images {
			match, err := regexp.MatchString(image, name)
			if err != nil {
				log.WithError(err).Warn("Image regexp not valid")
			}
			if match {
				return prefix.Prefix
			}
		}
	}

	return ""
}
