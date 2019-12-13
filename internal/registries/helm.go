package registries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

type Charts struct {
	Data []Chart `json:"data"`
}

type Chart struct {
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	Version string `json:"version"`
}

// GetLatestVersionFromHelm fetches latest version of the helm chart
func GetLatestVersionFromHelm(chart string) string {
	log.Infof("Fetching version for chart [%s]", chart)
	url := fmt.Sprintf("https://hub.helm.sh/api/chartsvc/v1/charts/stable/%s/versions", chart)
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Failed to fetch chart info [%v]", err)
		return versioning.Failure
	}

	defer resp.Body.Close()

	chartsData := Charts{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&chartsData)
	if err != nil {
		return versioning.Failure
	}
	var versions []string
	for _, data := range chartsData.Data {
		versions = append(versions, data.Attributes.Version)
	}
	return versioning.FindHigestVersionInList(versions)
}
