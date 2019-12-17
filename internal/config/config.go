package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/arminc/k8s-platform-lcm/internal/scanning"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

// Config of the lcm application, normally loaded from the config file
type Config struct {
	CliFlags               AppConfig
	AppConfig              AppConfig                  `koanf:"app"`
	KubernetesFetchEnabled bool                       `koanf:"kubernetesFetchEnabled"`
	Namespaces             []string                   `koanf:"namespaces"`
	ImageRegistries        registries.ImageRegistries `koanf:"imageRegistries"`
	ImageScanners          scanning.ImageScanners     `koanf:"imageScanners"`
	ToolRegistries         registries.ToolRegistries  `koanf:"toolRegistries"`
	Tools                  []registries.Tool          `koanf:"tools"`
	Images                 []string                   `koanf:"images"`
}

// AppConfig is the config for the app which can be set trough cli and config
type AppConfig struct {
	Locally bool
	Verbose bool `koanf:"verbose"`
	Debug   bool `koanf:"debug"`
}

// LoadConfiguration loads the configuration from file
func LoadConfiguration() Config {
	fileName := "config.yaml"
	log.WithField("configFile", fileName).Debug("Loading config file")

	var lcmConfig Config
	k := koanf.New(".")

	// load defaults
	if err := k.Load(confmap.Provider(map[string]interface{}{
		"kubernetesFetchEnabled": "true",
	}, "."), nil); err != nil {
		log.WithError(err).Fatal("Error loading config")
	}

	if err := k.Load(file.Provider(fileName), yaml.Parser()); err != nil {
		log.WithError(err).Fatal("Error loading config")
	}

	if err := k.Unmarshal("", &lcmConfig); err != nil {
		log.WithError(err).Fatal("Error unmarshaling config")
	}

	lcmConfig.ImageRegistries.DefaultRegistries()
	return lcmConfig
}

// IsVerboseLoggingEnabled returns true when verbose logging is enabled
func (c Config) IsVerboseLoggingEnabled() bool {
	return c.AppConfig.Verbose || c.CliFlags.Verbose
}

// IsDebugLoggingEnabled returns true when debug logging is enabled
func (c Config) IsDebugLoggingEnabled() bool {
	return c.AppConfig.Verbose || c.CliFlags.Verbose
}

// IsKubernetesFetchEnabled returns true when Kubernetes fetch is enabled
func (c Config) IsKubernetesFetchEnabled() bool {
	return c.KubernetesFetchEnabled
}

// RunningLocally returns true when running locally instead of in Kubernetes
func (c Config) RunningLocally() bool {
	return c.CliFlags.Locally
}
