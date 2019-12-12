package registries

import (
	log "github.com/sirupsen/logrus"
)

// ImageRegistries contains all the information regarding image registries
type ImageRegistries struct {
	Registries           []ImageRegistry    `koanf:"registries"`
	OverrideImages       []OverrideImage    `koanf:"override"`
	DefaultImageRegistry string             `koanf:"default"`
	OverrideRegistries   []OverrideRegistry `koanf:"overrideRegistries"`
}

// OverrideImage contains information about which registry to use, it overrides the url used in kubernetes
type OverrideImage struct {
	Name         string `koanf:"name"`
	RegistryName string `koanf:"registryName"`
}

// OverrideRegistry contains information about which registry to use, it overrides the url used in kubernetes
type OverrideRegistry struct {
	URL          string `koanf:"url"`
	RegistryName string `koanf:"registryName"`
}

// GetLatestVersionForImage gets the latest version for image
func (i ImageRegistries) GetLatestVersionForImage(name, url string) string {
	registry := i.determinRegistry(name, url)
	return registry.GetLatestVersion(name)
}

func (i ImageRegistries) determinRegistry(name, url string) ImageRegistry {
	registry, exists := i.FindRegistryByOverrideByImage(name)
	if exists {
		return registry
	}

	registry, exists = i.FindRegistryByOverrideByURL(url)
	if exists {
		return registry
	}

	registry, exists = i.GetDefaultRegistry()
	if exists {
		return registry
	}

	return i.FindRegistryByURL(url)
}

// FindRegistryByOverrideByImage finds if the image has an registry override
func (i ImageRegistries) FindRegistryByOverrideByImage(name string) (ImageRegistry, bool) {
	for _, image := range i.OverrideImages {
		if image.Name == name {
			for _, registry := range i.Registries {
				if registry.Name == image.RegistryName {
					return registry, true
				}
			}
		}
	}

	return ImageRegistry{}, false
}

// FindRegistryByOverrideByURL finds if the url has an registry override
func (i ImageRegistries) FindRegistryByOverrideByURL(url string) (ImageRegistry, bool) {
	for _, reg := range i.OverrideRegistries {
		if reg.URL == url {
			for _, registry := range i.Registries {
				if registry.Name == reg.RegistryName {
					return registry, true
				}
			}
		}
	}

	return ImageRegistry{}, false
}

// GetDefaultRegistry finds the default configured registry
func (i ImageRegistries) GetDefaultRegistry() (ImageRegistry, bool) {
	if i.DefaultImageRegistry != "" {
		for _, registry := range i.Registries {
			if registry.Name == i.DefaultImageRegistry {
				return registry, true
			}
		}
		log.Fatalf("Tried to find the default registry but it's not there")
	}
	log.Debugf("Default registry not set")
	return ImageRegistry{}, false
}

// FindRegistryByURL finds the configured registry by URL or creates an empty one based on the incoming URL with default settings
func (i ImageRegistries) FindRegistryByURL(url string) ImageRegistry {
	if url == "" {
		log.Debugf("We assume an empty registry means DockerHub")
		for _, registry := range i.Registries {
			if registry.Name == DockerHub {
				return registry
			}
		}
		log.Fatalf("Receive empty URL, assuming it's DockerHub but could not find DockerHub config")
	}

	for _, registry := range i.Registries {
		log.Debugf("Compare registry url [%s] with incoming url [%s]", registry.URL, url)
		if registry.URL == url {
			return registry
		}
	}

	log.Debugf("Could not find the registry, creating one with default information")
	return ImageRegistry{
		Name:     "Empty",
		URL:      url,
		AuthType: "none",
	}
}
