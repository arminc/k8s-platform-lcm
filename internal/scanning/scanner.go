package scanning

import (
	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/target/go-arty/xray"
)

const (
	// ERROR defines error for not able to fetch cve
	ERROR = "ERROR"
)

// GetVulnerabilities gets vulnerabilites for alle images using the configured scanner
func GetVulnerabilities(image kubernetes.Container, scanner config.ImageScanner) []string {
	log.Debug("Check if Vulnerabilitiy scanners are configured")
	//if !config.LcmConfig.AreScannersDefined() {
	//	return []string{}
	//}
	log.Debugf("Scan image: [%v]", image.Name)
	vul, err := getVulnerabilitiesFromXray(image, scanner)
	if err != nil {
		log.Errorf("Could not get vulnerabilities for [%s], error occured: [%v]", image.Name, err)
		return []string{ERROR}
	}
	return convertXrayToCves(vul, scanner)
}

func convertXrayToCves(artifacts []xray.SummaryArtifact, scanner config.ImageScanner) []string {
	cves := []string{}
	for _, issue := range artifacts[0].GetIssues() {
		log.Debugf("Issue: [%s]", issue.GetSummary())
		if scanner.IsSeverityEnabled(issue.GetSeverity()) && issue.GetSeverity() != "" {
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
