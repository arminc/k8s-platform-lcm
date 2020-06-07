package xray

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getConfig() Config {
	return Config{
		URL:      os.Getenv("xrayurl"),
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
		t.Skip("Skipping testing if no Xray configured")
	}
	client, _ := NewXray(getConfig())
	prefixes := []Prefix{
		{
			Prefix: os.Getenv("xrayprefix"),
			Images: []string{"awscli"},
		},
	}
	vul, err := client.GetVulnerabilities("awscli", "1.16.238-1", prefixes)
	assert.NoError(t, err, "No error expected")
	log.Info(vul)
	assert.Equal(t, 5, len(vul), "Should have 5 vulnerabilities")
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
