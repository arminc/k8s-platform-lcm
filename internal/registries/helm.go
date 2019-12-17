package registries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

// Charts is data structure coming from hub.helm.sh
type Charts struct {
	Data []Chart `json:"data"`
}

// Chart contains attribute information for a chart coming from hub.helm.sh
type Chart struct {
	Attributes Attributes `json:"attributes"`
}

// Attributes contains version information for a chart coming from hub.helm.sh
type Attributes struct {
	Version string `json:"version"`
}

// GetLatestVersionFromHelm fetches the latest version of the helm chart
func GetLatestVersionFromHelm(chart string) string {
	log.WithField("chart", chart).Debug("Fetching version for chart")
	url := fmt.Sprintf("https://hub.helm.sh/api/chartsvc/v1/charts/stable/%s/versions", chart)
	resp, err := http.Get(url)
	if err != nil {
		log.WithError(err).Error("Failed to fetch chart info")
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
	return versioning.FindHighestVersionInList(versions)
}
