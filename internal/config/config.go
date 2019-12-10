package config

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

const (
	// DockerHub is the default name for the DockerHub registry
	DockerHub = "DockerHub"
	// AuthTypeBasic is the basic auth type
	AuthTypeBasic = "basic"
	// AuthTypeToken is the token auth type
	AuthTypeToken = "token"
)

// Config of the lcm application, normally loaded from the config file
type Config struct {
	Namespaces           []string           `koanf:"namespaces"`
	ImageRegistries      []ImageRegistry    `koanf:"imageRegistries"`
	DefaultImageRegistry string             `koanf:"defaultImageRegistry"`
	OverrideImages       []OverrideImage    `koanf:"overrideImages"`
	OverrideRegistries   []OverrideRegistry `koanf:"overrideRegistries"`
	ImageScanners        []ImageScanner     `koanf:"imageScanners"`
}

// CommandFlags are flags to manipulate app behavior from the cli
type CommandFlags struct {
	LocalKubernetes bool
	Verbose         bool
	Debug           bool
}

// ImageRegistry contains all the information about the registry
type ImageRegistry struct {
	Name     string `koanf:"name"`
	URL      string `koanf:"url"`
	AuthType string `koanf:"authType"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
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

// ImageScanner contains all the information about the vulnerability scanner
type ImageScanner struct {
	Name     string            `koanf:"name"`
	URL      string            `koanf:"url"`
	Username string            `koanf:"username"`
	Password string            `koanf:"password"`
	Severity []string          `koanf:"severity"`
	Extra    map[string]string `koanf:"extra"`
}

// LcmConfig is singleton access to Config struct
var LcmConfig Config

// ConfigFlags is singleton access to Command flags
var ConfigFlags CommandFlags

// LoadConfiguration loads the configuration from file
func LoadConfiguration() {
	loadConfiguration("config.yaml")
}

// AreScannersDefined returns true if scanners are defined
func (c Config) AreScannersDefined() bool {
	if len(LcmConfig.ImageScanners) >= 1 {
		return true
	}
	return false
}

// IsSeverityEnabled checks if the severity is configured
func (i ImageScanner) IsSeverityEnabled(severity string) bool {
	for _, s := range i.Severity {
		if s == severity {
			return true
		}
	}
	return false
}

func loadConfiguration(fileName string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	k := koanf.New(".")
	if err := k.Load(file.Provider(fileName), yaml.Parser()); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	k.Unmarshal("", &LcmConfig)
}

// FindRegistryByOverrideByURL finds if the url has an registry override
func (c Config) FindRegistryByOverrideByURL(url string) (ImageRegistry, bool) {
	for _, reg := range c.OverrideRegistries {
		if reg.URL == url {
			for _, registry := range c.ImageRegistries {
				if registry.Name == reg.RegistryName {
					return registry, true
				}
			}
		}
	}

	return ImageRegistry{}, false
}

// FindRegistryByOverrideByImage finds if the image has an registry override
func (c Config) FindRegistryByOverrideByImage(name string) (ImageRegistry, bool) {
	for _, image := range c.OverrideImages {
		if image.Name == name {
			for _, registry := range c.ImageRegistries {
				if registry.Name == image.RegistryName {
					return registry, true
				}
			}
		}
	}

	return ImageRegistry{}, false
}

// GetDefaultRegistry finds the default configured registry
func (c Config) GetDefaultRegistry() (ImageRegistry, bool) {
	if c.DefaultImageRegistry != "" {
		for _, registry := range c.ImageRegistries {
			if registry.Name == c.DefaultImageRegistry {
				return registry, true
			}
		}
		log.Fatalf("Tried to find the default registry but it's not there")
	}
	log.Debugf("Default registry not set")
	return ImageRegistry{}, false
}

// FindRegistryByURL finds the configured registry by URL or creates an empty one based on the incoming URL with default settings
func (c Config) FindRegistryByURL(url string) ImageRegistry {
	if url == "" {
		log.Debugf("We assume an empty registry means DockerHub")
		for _, registry := range c.ImageRegistries {
			if registry.Name == DockerHub {
				return registry
			}
		}
		log.Fatalf("Receive empty URL, assuming it's DockerHub but could not find DockerHub config")
	}

	for _, registry := range c.ImageRegistries {
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
