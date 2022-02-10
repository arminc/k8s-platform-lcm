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
	Gitlab             ImageRegistry      `koanf:"gitlab"`
	OverrideImages     []OverrideImage    `koanf:"override"`
	OverrideRegistries []OverrideRegistry `koanf:"overrideRegistries"`
	OverrideImageNames map[string]string  `koanf:"overrideImageNames"`
}

// OverrideImage contains information about which registry to use, it overrides the URL used in kubernetes
type OverrideImage struct {
	Images           []string      `koanf:"images"`
	Registry         ImageRegistry `koanf:"registry"`
	RegistryName     string        `koanf:"registryName"`
	AllowAllReleases bool          `koanf:"allowAllReleases"`
}

// OverrideRegistry contains information about which registry to use, it overrides the URL used in kubernetes
type OverrideRegistry struct {
	Urls             []string      `koanf:"urls"`
	Registry         ImageRegistry `koanf:"registry"`
	RegistryName     string        `koanf:"registryName"`
	AllowAllReleases bool          `koanf:"allowAllReleases"`
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

	i.Gitlab.Name = Gitlab
	i.Gitlab.URL = "registry.gitlab.com"
	if i.Gitlab.AuthType == "" {
		i.Gitlab.AuthType = AuthTypeNone
	}
}

// GetLatestVersionForImage gets the latest version for image
func (i ImageRegistries) GetLatestVersionForImage(name, url string) string {
	registry := i.DetermineRegistry(name, url)
	name = i.findImageNameOverride(name)
	return registry.GetLatestVersion(name)
}

func (i ImageRegistries) DetermineRegistry(name, url string) ImageRegistry {
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

func (i ImageRegistries) findImageNameOverride(name string) string {
	overrideName := i.OverrideImageNames[name]
	if overrideName == "" {
		return name
	}
	return overrideName
}

// FindRegistryByOverrideByImage finds if the image has a registry override
func (i ImageRegistries) FindRegistryByOverrideByImage(name string) (ImageRegistry, bool) {
	for _, overrideImage := range i.OverrideImages {
		for _, image := range overrideImage.Images {
			match, err := regexp.MatchString(image, name)
			if err != nil {
				log.WithError(err).Fatal("Image regexp not valid")
			}
			if match {
				if overrideImage.RegistryName != "" {
					registry := i.FindRegistryByName(overrideImage.RegistryName)
					registry.AllowAllReleases = overrideImage.AllowAllReleases
					return registry, true
				}
				registry := overrideImage.Registry
				registry.AllowAllReleases = overrideImage.AllowAllReleases
				return registry, true
			}
		}
	}
	return ImageRegistry{}, false
}

// FindRegistryByOverrideByURL finds if the URL has a registry override
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
	} else if i.Gitlab.Default {
		return i.Gitlab, true
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
	} else if i.Gitlab.URL == url {
		return i.Gitlab
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
	} else if i.Gitlab.Name == name {
		return i.Gitlab
	}
	return i.DockerHub
}
