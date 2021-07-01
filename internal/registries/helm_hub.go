package registries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

// Chart contains attribute information for a chart coming from hub.helm.sh
type Chart struct {
	Packages []Packages `json:"packages"`
}

// Attributes contains version information for a chart coming from hub.helm.sh
type Packages struct {
	Version    string     `json:"version"`
	Repository Repository `json:"repository,omitempty"`
}

type Repository struct {
	Name string `json:"name"`
}

// SearchResultData contains search results from hub.helm.sh
type SearchResultData struct {
	Data []ChartSearchResult `json:"data"`
}

// ChartSearchResult contains chart search results from hub.helm.sh
type ChartSearchResult struct {
	ID string `json:"package_id"`
}

func (h HelmRegistries) useArtifactHub(chart string) string {
	repositoryName := ""
	chartName := chart

	overrideChartValue := h.OverrideChartNames[chart]
	overrideChart := strings.Split(overrideChartValue, "/")
	if len(overrideChart) > 1 {
		chartName = overrideChart[1]
		repositoryName = overrideChart[0]
	}

	version, err := findChartVersion(repositoryName, chartName)
	if err != nil {
		log.WithError(err).WithField("chart", chart).Error("Failed to search chart info")
		return versioning.Failure
	}

	return version
}

func findChartVersion(repositoryName, chartName string) (string, error) {
	repoParam := ""

	if repositoryName != "" {
		repoParam = fmt.Sprintf("&repo=%s", repositoryName)
	}
	// result is always limited to 1, because we can't handle multiple results anyway
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/search?facets=false&kind=0&deprecated=true&operators=false&verified_publisher=false&official=false&sort=stars&limit=5&ts_query_web=%s%s", chartName, repoParam)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	chartData := Chart{}
	err = json.NewDecoder(resp.Body).Decode(&chartData)

	if err != nil {
		return "", err
	}
	if len(chartData.Packages) == 0 {
		return "", fmt.Errorf("Could not find the chart")
	}
	if len(chartData.Packages) > 1 {

		sources := []string{}
		for _, key := range chartData.Packages {
			sources = append(sources, key.Repository.Name)
		}
		return "", fmt.Errorf("found more than one result, source repositories: %s, filter down with helmRegistries.overrideChartNames", sources)
	}

	log.Info(chartData.Packages)

	return chartData.Packages[0].Version, nil

}
