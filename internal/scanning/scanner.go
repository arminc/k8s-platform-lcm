package scanning

import (
	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
	"github.com/target/go-arty/xray"
)

// ImageScanners contains all the information about the vulnerability scanners
type ImageScanners struct {
	Severity []string   `koanf:"severity"`
	Xray     XrayConfig `koanf:"xray"`
}

// GetVulnerabilities gets vulnerabilities for all images using the configured scanner
func (i ImageScanners) GetVulnerabilities(name, version string) []string {
	if i.Xray.URL == "" {
		log.Debug("Xray not enabled")
		return []string{versioning.Nodata}
	}
	log.Debugf("Scan image: [%v]", name)
	vul, err := i.Xray.GetVulnerabilities(name, version)
	if err != nil {
		log.WithField("image", name).WithError(err).Error("Could not get vulnerabilities")
		return []string{versioning.Failure}
	}
	return i.convertXrayToCves(vul)
}

func (i ImageScanners) convertXrayToCves(artifacts []xray.SummaryArtifact) []string {
	cves := []string{}
	for _, issue := range artifacts[0].GetIssues() {
		log.WithField("summary", issue.GetSummary()).Debug("Issue")
		if i.isSeverityEnabled(issue.GetSeverity()) && issue.GetSeverity() != "" {
			for _, c := range issue.GetCves() {
				log.WithField("cve", c.GetCve()).Debug("CVE")
				cves = append(cves, c.GetCve())
			}
		} else {
			log.WithField("severity", issue.GetSeverity()).Debug("Severity not enabled")
		}
	}
	return cves
}

func (i ImageScanners) isSeverityEnabled(severity string) bool {
	for _, s := range i.Severity {
		if s == severity {
			return true
		}
	}
	return false
}
