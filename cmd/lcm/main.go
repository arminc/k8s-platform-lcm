package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/arminc/k8s-platform-lcm/internal"
	"github.com/arminc/k8s-platform-lcm/internal/config"
	log "github.com/sirupsen/logrus"
)

var (
	version = "0.1.0"
)

func initLogging(flags config.CommandFlags) {
	log.SetOutput(os.Stdout)     // Default to out instead of err
	log.SetLevel(log.ErrorLevel) // Default only Errors
	if flags.Verbose {
		log.SetLevel(log.InfoLevel)
	} else if flags.Debug {
		log.SetLevel(log.DebugLevel)
	}
}

func initFlags() config.CommandFlags {
	app := kingpin.New("lcm", "Kubernetes platform lifecycle management")
	app.Version(version)
	commandFlags := new(config.CommandFlags)
	app.Flag("local", "Run locally, default expected behavior is to run in the cluster").BoolVar(&commandFlags.LocalKubernetes)
	app.Flag("verbose", "Show more information").BoolVar(&commandFlags.Verbose)
	app.Flag("debug", "Show debug information, debug includes verbose").BoolVar(&commandFlags.Debug)
	app.Flag("nok8s", "Don't fetch data from kubernetes").BoolVar(&commandFlags.DisableKubernetesFetch)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	return *commandFlags
}

func main() {
	commandFlags := initFlags()
	config := config.LoadConfiguration()
	config.CommandFlags = commandFlags
	initLogging(config.CommandFlags)
	log.Infof("Running version %s", version)
	internal.Execute(config)
}
