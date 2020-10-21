package registries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

type Charts struct {
	Data Chart `json:"data"`
}

// Chart contains attribute information for a chart coming from hub.helm.sh
type Chart struct {
	Packages []Packages `json:"packages"`
}

// Attributes contains version information for a chart coming from hub.helm.sh
type Packages struct {
	Version string `json:"version"`
}

// SearchResultData contains search results from hub.helm.sh
type SearchResultData struct {
	Data []ChartSearchResult `json:"data"`
}

// ChartSearchResult contains chart search results from hub.helm.sh
type ChartSearchResult struct {
	ID string `json:"package_id"`
}

func (h HelmRegistries) useHelmHub(chart string) string {
	chartName := h.OverrideChartNames[chart]
	chartName = strings.Replace(chartName, "/", "%20", -1)
	if chartName == "" {
		chartName = chart
	}
	version, err := findChartVersion(chartName)
	if err != nil {
		log.WithError(err).WithField("chart", chart).Error("Failed to search chart info")
		return versioning.Failure
	}

	return version
}

func findChartVersion(chartName string) (string, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/search?limit=1&facets=false&ts_query_web=%s&kind=0", chartName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	chartData := Charts{}
	err = json.NewDecoder(resp.Body).Decode(&chartData)

	if err != nil {
		return "", err
	}
	if len(chartData.Data.Packages) == 0 {
		return "", fmt.Errorf("Could not find the chart")
	}
	log.Info(chartData.Data.Packages)

	return chartData.Data.Packages[0].Version, nil

}
