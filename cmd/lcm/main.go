package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/arminc/k8s-platform-lcm/internal"
	"github.com/arminc/k8s-platform-lcm/internal/config"
	log "github.com/sirupsen/logrus"
)

// Version is the current app version
var Version = "dev"

func initLogging(config config.Config) {
	log.SetOutput(os.Stdout)     // Default to out instead of err
	log.SetLevel(log.ErrorLevel) // Default only Errors
	if config.IsVerboseLoggingEnabled() {
		log.SetLevel(log.InfoLevel)
	} else if config.IsDebugLoggingEnabled() {
		log.SetLevel(log.DebugLevel)
	}

	if config.IsJSONLoggingEnabled() {
		log.SetFormatter(&log.JSONFormatter{})
	}
	enabled, logFile := config.LogToFilePath()
	if enabled {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Fatal("Could not log to file")
		}
		log.SetOutput(file)
	}
}

func initFlags() config.AppConfig {
	app := kingpin.New("lcm", "Kubernetes platform lifecycle management")
	app.Version(Version)
	cliFlags := new(config.AppConfig)
	app.Flag("config", "Provide the path to the config file. Default is config.yaml which is in the same folder as lcm").Default("config.yaml").StringVar(&cliFlags.ConfigFile)
	app.Flag("local", "Run locally, default expected behavior is to run in the Kubernetes cluster").BoolVar(&cliFlags.Locally)
	app.Flag("verbose", "Show more information. This overrides the config setting").BoolVar(&cliFlags.Verbose)
	app.Flag("debug", "Show debug information, debug includes verbose. This overrides the config setting").BoolVar(&cliFlags.Debug)
	app.Flag("jsonLogging", "Log in json format").BoolVar(&cliFlags.JSONLoggingEnabled)
	app.Flag("logFile", "Log file path").StringVar(&cliFlags.LogFile)
	app.Flag("server", "Start the server").BoolVar(&cliFlags.StartServer)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	return *cliFlags
}

func main() {
	cliFlags := initFlags()
	config := config.LoadConfiguration(cliFlags.ConfigFile)
	config.CliFlags = cliFlags // Add cli flags to config object
	initLogging(config)
	log.WithField("version", Version).Info("Running version")
	internal.Execute(config)
	if config.CliFlags.StartServer {
		internal.StartServer()
	}
}
