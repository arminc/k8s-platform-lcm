package scanning

import (
	log "github.com/sirupsen/logrus"
	"github.com/target/go-arty/xray"
)

const (
	// ERROR defines error for not able to fetch cve
	ERROR = "ERROR"
)

// ImageScanners contains all the information about the vulnerability scanners
type ImageScanners struct {
	Severity []string   `koanf:"severity"`
	Xray     XrayConfig `koanf:"xray"`
}

// GetVulnerabilities gets vulnerabilites for alle images using the configured scanner
func (i ImageScanners) GetVulnerabilities(name, version string) []string {
	if i.Xray.URL == "" {
		log.Debug("Xray not enabled")
		return nil
	}
	log.Debugf("Scan image: [%v]", name)
	vul, err := i.Xray.GetVulnerabilities(name, version)
	if err != nil {
		log.Errorf("Could not get vulnerabilities for [%s], error occured: [%v]", name, err)
		return []string{ERROR}
	}
	return i.convertXrayToCves(vul)
}

func (i ImageScanners) convertXrayToCves(artifacts []xray.SummaryArtifact) []string {
	cves := []string{}
	for _, issue := range artifacts[0].GetIssues() {
		log.Debugf("Issue: [%s]", issue.GetSummary())
		if i.IsSeverityEnabled(issue.GetSeverity()) && issue.GetSeverity() != "" {
			for _, c := range issue.GetCves() {
				log.Debugf("CVE: [%s]", c.GetCve())
				cves = append(cves, c.GetCve())
			}
		} else {
			log.Infof("Severity not enabled: [%s]", issue.GetSeverity())
		}
	}
	return cves
}

// IsSeverityEnabled checks if the severity is configured
func (i ImageScanners) IsSeverityEnabled(severity string) bool {
	for _, s := range i.Severity {
		if s == severity {
			return true
		}
	}
	return false
}
