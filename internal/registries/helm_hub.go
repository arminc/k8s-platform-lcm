package registries

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

// ChartInfo is data structure coming from artifacthub.io
type ChartInfo struct {
	AvailableVersions []AvailableVersions `json:"available_versions"`
}

// AvailableVersions contains version information for a chart coming from artifacthub.io
type AvailableVersions struct {
	Version   string `json:"version"`
	CreatedAt int    `json:"created_at"`
}

// SearchResultData contains search results from artifacthub.io
type SearchResultData struct {
	Data []ChartSearchResult `json:"data"`
}

// ChartSearchResult contains chart search results from artifacthub.io
type ChartSearchResult struct {
	ID string `json:"id"`
}

func (h HelmRegistries) useHelmHub(chart string) string {
	chartName := h.OverrideChartNames[chart]
	if chartName == "" {
		var err error
		chartName, err = findChart(chart)
		if err != nil {
			log.WithError(err).WithField("chart", chart).Error("Failed to search chart info")
			return versioning.Failure
		}
	}

	versions, err := getChartVersions(chartName)
	if err != nil {
		log.WithError(err).WithField("chart", chart).Error("Failed to fetch chart info")
		return versioning.Failure
	}

	return versioning.FindHighestVersionInList(versions, false)
}

func findChart(chart string) (string, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/chartsvc/v1/charts/search?q=%s", url.QueryEscape(chart))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	searchData := SearchResultData{}
	err = json.NewDecoder(resp.Body).Decode(&searchData)
	if err != nil {
		return "", err
	}

	if len(searchData.Data) == 0 {
		return "", fmt.Errorf("Could not find the chart")
	} else if len(searchData.Data) == 1 {
		return searchData.Data[0].ID, nil
	}
	return "", fmt.Errorf("More than one result %v", searchData)
}

func getChartVersions(chart string) ([]string, error) {
	url := fmt.Sprintf("https://artifacthub.io/api/v1/packages/helm/%s", chart)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	chartInfo := ChartInfo{}
	err = json.NewDecoder(resp.Body).Decode(&chartInfo)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, version := range chartInfo.AvailableVersions {
		versions = append(versions, version.Version)
	}
	return versions, nil
}
