package xray

import (
	"fmt"
	"regexp"

	"github.com/arminc/k8s-platform-lcm/pkg/vulnerabilities"
	log "github.com/sirupsen/logrus"
	"github.com/target/go-arty/xray"
)

// Config contains all the information to talk to Xray
type Config struct {
	Username string   `koanf:"username"`
	Password string   `koanf:"password"`
	Url      string   `koanf:"url"`
	Prefixes []Prefix `koanf:"prefixes"`
}

// Prefix information about the index used by Xray
type Prefix struct {
	Prefix string   `koanf:"prefix"`
	Images []string `koanf:"images"`
}

type XrayScanner interface {
	GetVulnerabilities(name, version string, prefixes []Prefix) ([]vulnerabilities.Vulnerability, error)
	GetXrayResults(request xray.SummaryArtifactRequest) (xray.SummaryArtifact, error)
}

type xrayClient struct {
	client *xray.Client
}

func NewXray(config Config) (XrayScanner, error) {
	client, err := xray.NewClient(config.Url, nil)
	if err != nil {
		return &xrayClient{}, err
	}
	client.Authentication.SetBasicAuth(config.Username, config.Password)
	return &xrayClient{
		client: client,
	}, nil
}

func (x *xrayClient) GetVulnerabilities(name, version string, prefixes []Prefix) ([]vulnerabilities.Vulnerability, error) {
	path := fmt.Sprintf("%s/%s/%s", findPrefix(name, prefixes), name, version)
	xrayVulnerabilities, err := x.GetXrayResults(xray.SummaryArtifactRequest{
		Paths: &[]string{path},
	})
	if err != nil {
		return nil, err
	}

	var allVulnerabilities []vulnerabilities.Vulnerability
	for _, issue := range xrayVulnerabilities.GetIssues() {
		for _, cve := range issue.GetCves() {
			vulnerability := vulnerabilities.Vulnerability{
				Identifier:  cve.GetCve(),
				Description: *issue.Description,
			}
			allVulnerabilities = append(allVulnerabilities, vulnerability)
		}
	}
	return allVulnerabilities, nil
}

func (x *xrayClient) GetXrayResults(request xray.SummaryArtifactRequest) (xray.SummaryArtifact, error) {
	sum, response, err := x.client.Summary.Artifact(&request)
	if err != nil {
		return xray.SummaryArtifact{}, err
	}
	if response.StatusCode != 200 {
		log.WithField("request", request).Warnf("Error fetching xray vulnerabilities: http-status: %s", response.Status)
		return xray.SummaryArtifact{}, fmt.Errorf("Error fetching xray vulnerabilities, http-status: %s", response.Status)
	}
	if len(sum.GetErrors()) >= 1 {
		return xray.SummaryArtifact{}, fmt.Errorf("Got an error from xray for [%v] error [%s]", request, *sum.GetErrors()[0].Error)
	}
	if len(sum.GetArtifacts()) > 0 {
		return sum.GetArtifacts()[0], nil
	}
	return xray.SummaryArtifact{}, nil
}

// findPrefix returns the prefix used by Xray for the image
func findPrefix(imageName string, prefixes []Prefix) string {
	if len(prefixes) == 1 {
		return prefixes[0].Prefix
	}
	for _, prefix := range prefixes {
		for _, image := range prefix.Images {
			match, err := regexp.MatchString(image, imageName)
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
