package registries

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func registry(name, url, authtype, username, password, region string, is_default, allow_all_releases bool) ImageRegistry {
	return ImageRegistry{
		Name:             name,
		URL:              url,
		AuthType:         authtype,
		Username:         username,
		Password:         password,
		Region:           region,
		Default:          is_default,
		AllowAllReleases: allow_all_releases,
	}
}

func TestECRAuth(t *testing.T) {
	registry := registry("test", "187113339385.dkr.ecr.us-east-1.amazonaws.com", AuthTypeECR, "", "", "", false, true)
	registry.GetLatestVersion("bitnami/minideb")
	version := registry.GetLatestVersion("alpine")

	expected := "3.14.2"
	assert.Equal(t, expected, version, "No version")
}
