package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

// Config of the lcm application, normally loaded from the config file
type Config struct {
	Namespaces      []string                   `koanf:"namespaces"`
	ImageRegistries registries.ImageRegistries `koanf:"images"`
	ImageScanners   []ImageScanner             `koanf:"imageScanners"`
	ToolRegistries  registries.ToolRegistries  `koanf:"toolRegistries"`
	Tools           []registries.Tool          `koanf:"tools"`
}

// CommandFlags are flags to manipulate app behavior from the cli
type CommandFlags struct {
	LocalKubernetes bool
	Verbose         bool
	Debug           bool
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

// ConfigFlags is singleton access to Command flags
var ConfigFlags CommandFlags

// LoadConfiguration loads the configuration from file
func LoadConfiguration() Config {
	fileName := "config.yaml"

	var lcmConfig Config
	k := koanf.New(".")
	if err := k.Load(file.Provider(fileName), yaml.Parser()); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	err := k.Unmarshal("", &lcmConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}
	return lcmConfig
}

// AreScannersDefined returns true if scanners are defined
func (c Config) AreScannersDefined() bool {
	return len(c.ImageScanners) >= 1
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
