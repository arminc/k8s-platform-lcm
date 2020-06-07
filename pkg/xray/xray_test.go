package xray

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/go-arty/xray"
)

func getConfig() Config {
	return Config{
		Url:      os.Getenv("xrayurl"),
		Username: os.Getenv("xrayusername"),
		Password: os.Getenv("xraypassword"),
	}
}

func TestConfigWrong(t *testing.T) {
	_, err := NewXray(Config{})
	assert.Error(t, err)
}

func TestGetVulnerability(t *testing.T) {
	if os.Getenv("xrayurl") == "" {
		t.Skip("Skipping testing in CI environment")
	}
	client, _ := NewXray(getConfig())
	vul, err := client.GetXrayResults(xray.SummaryArtifactRequest{
		Paths: &[]string{fmt.Sprintf("%s/%s/%s", os.Getenv("xrayprefix"), "alpine", "3.10")},
	})
	assert.NoError(t, err, "No error espected")
	assert.Equal(t, 1, len(vul), "Should have 1 vulnerabilities")
}

func TestPrefixNoMatchOne(t *testing.T) {
	prefix := findPrefix("alpine", []Prefix{})
	assert.Equal(t, "", prefix, "Should be empty")
}

func TestPrefixReturnFirstIfOnePrefix(t *testing.T) {
	prefix := findPrefix("alpine", []Prefix{
		{
			Prefix: "test",
			Images: []string{"ubuntu"},
		},
	})
	assert.Equal(t, "test", prefix, "Should be empty")
}

func TestPrefixNoMatchTwo(t *testing.T) {
	prefix := findPrefix("alpine", []Prefix{
		{
			Prefix: "test",
			Images: []string{"ubuntu"},
		},
		{
			Prefix: "some",
			Images: []string{"debian"},
		},
	})
	assert.Equal(t, "", prefix, "Should be empty")
}

func TestPrefixMatch(t *testing.T) {
	prefix := findPrefix("alpine", []Prefix{
		{
			Prefix: "test",
			Images: []string{"ubuntu"},
		},
		{
			Prefix: "some",
			Images: []string{"alpine"},
		},
	})
	assert.Equal(t, "some", prefix, "Should be empty")
}
