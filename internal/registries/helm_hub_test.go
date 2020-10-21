package registries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmSearch(t *testing.T) {
	chartId, err := findChart("alertmanager prometheus-community")
	assert.NoError(t, err, "No error expected")
	assert.Equal(t, "prometheus-community/alertmanager", chartId, "ID is not chartId")
}

func TestHelmVersion(t *testing.T) {
	chartIds, err := getChartVersions("prometheus-community/alertmanager")
	assert.NoError(t, err, "No error expected")
	assert.GreaterOrEqual(t, len(chartIds), 1, "Should at least have one id")
}
