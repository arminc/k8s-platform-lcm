// Package trivy is used to access trivy server to find vulnerabilities for images
package trivy

import (
	"crypto/sha1"
	"encoding/base64"

	"github.com/aquasecurity/trivy/pkg/report"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	"github.com/arminc/k8s-platform-lcm/pkg/vulnerabilities"
	log "github.com/sirupsen/logrus"
)

// Config contains all the information to talk to Trivy
type Config struct {
	URL string `koanf:"url"`
}

// Scanner is an interface that wraps calls to Xray
type Scanner interface {
	GetVulnerabilities(fullPath string) ([]vulnerabilities.Vulnerability, error)
	GetResults(request string) (report.Results, error)
}

type TrivyClient struct {
	scanner *client.Scanner
	url     string
}

// NewTrivy constructs a new Trivy client
// It returns an implementation of the Trivy client represented as the Scanner interface
func NewTrivy(config Config) (Scanner, error) {
	scanner, err := NewClient(config.URL, "")
	if err != nil {
		log.Debugf("Trivy error: %v", err)
		return &TrivyClient{}, err
	}

	return &TrivyClient{
		scanner: scanner,
		url:     config.URL,
	}, nil
}

// GetVulnerabilities returns Trivy results as generic Vulnerabilities instead of in the Trivy format
// It returns empty Image list on error
func (t *TrivyClient) GetVulnerabilities(fullPath string) ([]vulnerabilities.Vulnerability, error) {
	log.WithField("fullPath", fullPath).Debug("Fetching vulnerabilities")
	image := fullPath
	trivyVulnerabilities, err := t.GetResults(image)
	if err != nil {
		return []vulnerabilities.Vulnerability{}, err
	}

	allVulnerabilities := []vulnerabilities.Vulnerability{}

	for _, result := range trivyVulnerabilities {
		for _, vuln := range result.Vulnerabilities {
			cve := vuln.VulnerabilityID
			if cve == "" {
				cve = hashString(vuln.Description)
			}
			vulnerability := vulnerabilities.Vulnerability{
				Identifier:  cve,
				Description: vuln.Description,
				Severity:    vuln.Severity,
			}
			allVulnerabilities = append(allVulnerabilities, vulnerability)
		}
	}
	return allVulnerabilities, nil
}

// GetResults returns results as they come from Trivy
func (t *TrivyClient) GetResults(image string) (r report.Results, err error) {
	report, err := Run(t.scanner, t.url, image)

	if err != nil {
		return nil, err
	}

	severityCount := map[string]int{}
	for _, result := range (*report).Results {
		log.Debugf("%v, %v", result.Target, result.Type)
		for _, v := range result.Vulnerabilities {
			severityCount[v.Severity]++
			log.Debugf("%v", []string{v.PkgName, v.VulnerabilityID, v.Severity, v.InstalledVersion, v.FixedVersion})
		}
	}
	log.Warnf("%v", severityCount)

	return (*report).Results, nil
}

func hashString(text string) string {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		log.WithError(err).Warn("Could not hash")
		return "HASH_ERROR"
	}
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
