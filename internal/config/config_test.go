package config

import "testing"

func TestLoadingConfigFile(t *testing.T) {
	loadConfiguration("../../test/exampleConfig.yaml")

	namespacesLength := len(LcmConfig.Namespaces)
	if namespacesLength != 2 {
		t.Errorf("Expecting two namespaces got [%v]", namespacesLength)
	}

	imageRegistriesLength := len(LcmConfig.ImageRegistries)
	if imageRegistriesLength != 5 {
		t.Errorf("Expecting five registries got [%v]", imageRegistriesLength)
	}

	dockerHub := LcmConfig.ImageRegistries[0]
	if dockerHub.Name != DockerHub {
		t.Errorf("DockerHub registry Name should be [%s] but got [%s]", DockerHub, dockerHub.Name)
	}
	if dockerHub.URL != "registry.hub.docker.com" {
		t.Errorf("DockerHub registry Url should be [registry.hub.docker.com] but got [%s]", dockerHub.URL)
	}
	if dockerHub.AuthType != "token" {
		t.Errorf("DockerHub registry Token should be [token] but got [%s]", dockerHub.AuthType)
	}
	if dockerHub.Username != "" {
		t.Errorf("DockerHub registry Username should be [empty] but got [%s]", dockerHub.Username)
	}
	if dockerHub.Password != "" {
		t.Errorf("DockerHub registry Password should be [empty] but got [%s]", dockerHub.Password)
	}

	overrideImagesLength := len(LcmConfig.OverrideImages)
	if overrideImagesLength != 1 {
		t.Errorf("Expecting one image override got [%v]", overrideImagesLength)
	}
}

func TestFindRegistryEmptyUrl(t *testing.T) {
	loadConfiguration("../../test/exampleConfig.yaml")

	registry := LcmConfig.FindRegistryByURL("")
	if registry.Name != DockerHub {
		t.Errorf("Expected to get DockerHub registry but got [%v]", registry)
	}
}

func TestFindRegistryKnownUrl(t *testing.T) {
	loadConfiguration("../../test/exampleConfig.yaml")

	registry := LcmConfig.FindRegistryByURL("gcr.io")
	if registry.Name != "Gcr" {
		t.Errorf("Expected to get Gcr registry but got [%v]", registry)
	}
}

func TestFindRegistryUnknownUrl(t *testing.T) {
	loadConfiguration("../../test/exampleConfig.yaml")

	registry := LcmConfig.FindRegistryByURL("not.known.io")
	if registry.Name != "Empty" {
		t.Errorf("Expected to get Empty registry but got [%v]", registry)
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	loadConfiguration("../../test/exampleConfig.yaml")

	registry, exists := LcmConfig.GetDefaultRegistry()
	if !exists {
		t.Errorf("Expected default registry to exist")
	}
	if registry.URL != "private.somenonexistingurl.io" {
		t.Errorf("Received wrong default registry [%v]", registry)
	}
}
