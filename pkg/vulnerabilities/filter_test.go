package vulnerabilities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterSeverityNone(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
	}

	response := filterAcceptedSeverities(vul, []string{})
	assert.Equal(t, 1, len(response), "Should return vulnerabilities")
}

func TestFilterSeveritiesNoMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
	}

	response := filterAcceptedSeverities(vul, []string{"LOW", "INFO", "DEBUG"})
	assert.Equal(t, 1, len(response), "Should return vulnerabilities")
}

func TestFilterSeveritiesMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
	}

	response := filterAcceptedSeverities(vul, []string{"INFO", "HIGH", "DEBUG"})
	assert.Equal(t, 0, len(response), "Should return no vulnerabilities")
}

func TestFilterSeveritiesMatchMultipleVulnerabilities(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
		{
			Severity: "INFO",
		},
		{
			Severity: "CRITICAL",
		},
		{
			Severity: "DEBUG",
		},
		{
			Severity: "OTHER",
		},
	}

	response := filterAcceptedSeverities(vul, []string{"INFO", "HIGH", "DEBUG"})
	assert.Equal(t, 2, len(response), "Should return no vulnerabilities")
}
