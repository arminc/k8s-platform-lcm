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

	filter := FilterData{
		Severities: []string{},
	}

	response := filter.Vulnerabilities("", vul)
	assert.Equal(t, 1, len(response), "Should return vulnerabilities")
}

func TestFilterSeveritiesNoMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
	}

	filter := FilterData{
		Severities: []string{"LOW", "INFO", "DEBUG"},
	}

	response := filter.Vulnerabilities("", vul)
	assert.Equal(t, 1, len(response), "Should return vulnerabilities")
}

func TestFilterSeveritiesMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Severity: "HIGH",
		},
	}

	filter := FilterData{
		Severities: []string{"INFO", "HIGH", "DEBUG"},
	}

	response := filter.Vulnerabilities("", vul)
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

	filter := FilterData{
		Severities: []string{"INFO", "HIGH", "DEBUG"},
	}

	response := filter.Vulnerabilities("", vul)
	assert.Equal(t, 2, len(response), "Should return no vulnerabilities")
}

func TestFilterIdentifierNoNameMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Identifier: "CVE1",
		},
	}

	filter := FilterData{
		Identifiers: []Identifier{
			{
				Match:       "test",
				Identifiers: []string{"CVE1"},
			},
		},
	}

	response := filter.Vulnerabilities("wrong", vul)
	assert.Equal(t, 1, len(response), "Should return vulnerabilities")
}

func TestFilterIdentifierFullMatch(t *testing.T) {
	vul := []Vulnerability{
		{
			Identifier: "CVE1",
		},
	}

	filter := FilterData{
		Identifiers: []Identifier{
			{
				Match:       "test",
				Identifiers: []string{"CVE1"},
			},
		},
	}

	response := filter.Vulnerabilities("test", vul)
	assert.Equal(t, 0, len(response), "Should return vulnerabilities")
}

func TestFilterIdentifierFullRegExp(t *testing.T) {
	vul := []Vulnerability{
		{
			Identifier: "CVE1",
		},
	}

	filter := FilterData{
		Identifiers: []Identifier{
			{
				Match:       "te.*",
				Identifiers: []string{"CVE1"},
			},
		},
	}

	response := filter.Vulnerabilities("test", vul)
	assert.Equal(t, 0, len(response), "Should return vulnerabilities")
}

func TestFilterIdentifierWithTag(t *testing.T) {
	vul := []Vulnerability{
		{
			Identifier: "CVE1",
		},
	}

	filter := FilterData{
		Identifiers: []Identifier{
			{
				Match:       "arminc/some:1.2.1",
				Identifiers: []string{"CVE1"},
			},
		},
	}

	response := filter.Vulnerabilities("arminc/some:1.2.1", vul)
	assert.Equal(t, 0, len(response), "Should return vulnerabilities")
}
