package scanning

import (
	"fmt"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	"github.com/target/go-arty/xray"
)

func getVulnerabilitiesFromXray(image kubernetes.Container, scanner config.ImageScanner) ([]xray.SummaryArtifact, error) {
	url := "https://" + scanner.URL
	client, _ := xray.NewClient(url, nil)

	prefix := scanner.Extra["prefix"]
	path := fmt.Sprintf("%s/%s/%s", prefix, image.Name, image.Version)
	arty := &xray.SummaryArtifactRequest{
		Paths: &[]string{path},
	}
	client.Authentication.SetBasicAuth(scanner.Username, scanner.Password)
	sum, res, err := client.Summary.Artifact(arty)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Response code wrong [%v]", res.StatusCode)
	}
	if len(sum.GetErrors()) >= 1 {
		return nil, fmt.Errorf("Got an error from xray for [%s]m error [%s]", image.Name, *sum.GetErrors()[0].Error)
	}
	return sum.GetArtifacts(), nil
}
