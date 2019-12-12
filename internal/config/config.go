package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/arminc/k8s-platform-lcm/internal/scanning"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

// Config of the lcm application, normally loaded from the config file
type Config struct {
	CommandFlags    CommandFlags
	Namespaces      []string                   `koanf:"namespaces"`
	ImageRegistries registries.ImageRegistries `koanf:"imageRegistries"`
	ImageScanners   scanning.ImageScanners     `koanf:"imageScanners"`
	ToolRegistries  registries.ToolRegistries  `koanf:"toolRegistries"`
	Tools           []registries.Tool          `koanf:"tools"`
}

// CommandFlags are flags to manipulate app behavior from the cli
type CommandFlags struct {
	LocalKubernetes bool
	Verbose         bool
	Debug           bool
}

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
