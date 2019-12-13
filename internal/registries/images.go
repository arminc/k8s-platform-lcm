package registries

import (
	log "github.com/sirupsen/logrus"
	"regexp"
)

// ImageRegistries contains all the information regarding image registries
type ImageRegistries struct {
	DockerHub          ImageRegistry      `koanf:"dockerHub"`
	Quay               ImageRegistry      `koanf:"quay"`
	Gcr                ImageRegistry      `koanf:"gcr"`
	GcrK8s             ImageRegistry      `koanf:"gcrK8s"`
	Zalando            ImageRegistry      `koanf:"zalando"`
	OverrideImages     []OverrideImage    `koanf:"override"`
	OverrideRegistries []OverrideRegistry `koanf:"overrideRegistries"`
}

// OverrideImage contains information about which registry to use, it overrides the url used in kubernetes
type OverrideImage struct {
	Images       []string      `koanf:"images"`
	Registry     ImageRegistry `koanf:"registry"`
	RegistryName string        `koanf:"registryName"`
}

// OverrideRegistry contains information about which registry to use, it overrides the url used in kubernetes
type OverrideRegistry struct {
	Urls         []string      `koanf:"urls"`
	Registry     ImageRegistry `koanf:"registry"`
	RegistryName string        `koanf:"registryName"`
}

// DefaultRegistries sets default values for registries
func (i *ImageRegistries) DefaultRegistries() {
	i.DockerHub.Name = DockerHub
	i.DockerHub.URL = "registry.hub.docker.com"
	i.DockerHub.AuthType = AuthTypeToken

	i.Quay.Name = Quay
	i.Quay.URL = "quay.io"
	if i.Quay.AuthType == "" {
		i.Quay.AuthType = AuthTypeNone
	}

	i.Gcr.Name = Gcr
	i.Gcr.URL = "gcr.io"
	if i.Gcr.AuthType == "" {
		i.Gcr.AuthType = AuthTypeNone
	}

	i.GcrK8s.Name = GcrK8s
	i.GcrK8s.URL = "k8s.gcr.io"
	if i.GcrK8s.AuthType == "" {
		i.GcrK8s.AuthType = AuthTypeNone
	}

	i.Zalando.Name = Zalando
	i.Zalando.URL = "registry.opensource.zalan.do"
	if i.Zalando.AuthType == "" {
		i.Zalando.AuthType = AuthTypeNone
	}
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
	for _, overrideImage := range i.OverrideImages {
		for _, image := range overrideImage.Images {
			match, err := regexp.MatchString(image, name)
			if err != nil {
				log.Fatalf("Image regexp not valid [%v]", err)
			}
			log.Debugf("FindRegistryByOverrideByImage name [%s], image [%s], bool [%v]", name, image, match)
			if match {
				if overrideImage.RegistryName != "" {
					return i.FindRegistryByName(overrideImage.RegistryName), true
				}
				return overrideImage.Registry, true
			}
		}
	}
	return ImageRegistry{}, false
}

// FindRegistryByOverrideByURL finds if the url has an registry override
func (i ImageRegistries) FindRegistryByOverrideByURL(url string) (ImageRegistry, bool) {
	for _, overrideRegistry := range i.OverrideRegistries {
		for _, regURL := range overrideRegistry.Urls {
			if regURL == url {
				if overrideRegistry.RegistryName != "" {
					return i.FindRegistryByName(overrideRegistry.RegistryName), true
				}
				return overrideRegistry.Registry, true
			}
		}
	}
	return ImageRegistry{}, false
}

// GetDefaultRegistry finds the default configured registry
func (i ImageRegistries) GetDefaultRegistry() (ImageRegistry, bool) {
	if i.Quay.Default {
		return i.Quay, true
	} else if i.Gcr.Default {
		return i.Gcr, true
	} else if i.GcrK8s.Default {
		return i.GcrK8s, true
	} else if i.Zalando.Default {
		return i.Zalando, true
	} else if i.DockerHub.Default {
		return i.DockerHub, true
	}
	return ImageRegistry{}, false
}

// FindRegistryByURL finds the configured registry by URL, default is DockerHub
func (i ImageRegistries) FindRegistryByURL(url string) ImageRegistry {
	if i.Quay.URL == url {
		return i.Quay
	} else if i.Gcr.URL == url {
		return i.Gcr
	} else if i.GcrK8s.URL == url {
		return i.GcrK8s
	} else if i.Zalando.URL == url {
		return i.Zalando
	}
	return i.DockerHub
}

// FindRegistryByName finds the configured registry by name, default is DockerHub
func (i ImageRegistries) FindRegistryByName(name string) ImageRegistry {
	if i.Quay.Name == name {
		return i.Quay
	} else if i.Gcr.Name == name {
		return i.Gcr
	} else if i.GcrK8s.Name == name {
		return i.GcrK8s
	} else if i.Zalando.Name == name {
		return i.Zalando
	}
	return i.DockerHub
}
