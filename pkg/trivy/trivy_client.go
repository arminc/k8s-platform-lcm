package trivy

import (
	"context"
	"net/http"
	"time"

	"github.com/aquasecurity/fanal/analyzer"
	"github.com/aquasecurity/fanal/analyzer/config"
	"github.com/aquasecurity/fanal/artifact"
	image2 "github.com/aquasecurity/fanal/artifact/image"
	"github.com/aquasecurity/fanal/image"
	"github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/commands/operation"
	pkgReport "github.com/aquasecurity/trivy/pkg/report"
	"github.com/aquasecurity/trivy/pkg/rpc/client"
	scanner2 "github.com/aquasecurity/trivy/pkg/scanner"
	"github.com/aquasecurity/trivy/pkg/types"
	log "github.com/sirupsen/logrus"
)

const defaultPolicyNamespace = "appshield"

func disabledAnalyzers() []analyzer.Type {
	// Specified analyzers to be disabled depending on scanning modes
	// e.g. The 'image' subcommand should disable the lock file scanning.
	analyzers := analyzer.TypeLockfiles

	// It doesn't analyze apk commands by default.
	if true {
		analyzers = append(analyzers, analyzer.TypeApkCommand)
	}

	/*
		// Don't analyze programming language packages when not running in 'library' mode
		if !utils.StringInSlice(types.VulnTypeLibrary, opt.VulnType) {
			analyzers = append(analyzers, analyzer.TypeLanguages...)
		}
	*/

	return analyzers
}

func NewClient(remoteUrl string, placeholder string) (s *client.Scanner, err error) {
	log.Debug("Getting trivy client")

	url := client.RemoteURL(remoteUrl)
	customHeaders := client.CustomHeaders(make(map[string][]string))

	scannerScanner := client.NewProtobufClient(url)
	clientScanner := client.NewScanner(customHeaders, scannerScanner)

	return &clientScanner, nil
}

func Run(scanner *client.Scanner, url string, imageName string) (r *pkgReport.Report, err error) {
	ctx := context.Background()

	customCacheHeaders := http.Header(make(map[string][]string))

	remoteCache := cache.NewRemoteCache(cache.RemoteURL(url), customCacheHeaders)

	dockerOption, err := types.GetDockerOption(30 * time.Second)
	if err != nil {
		log.Debugf("trivy docker error: %v\n%v", err, dockerOption)
	}

	typesImage, cleanup, err := image.NewDockerImage(ctx, imageName, dockerOption)
	if err != nil {
		log.Debugf("trivy typesImage error: %v\n%v", err, typesImage)
		return nil, err
	}

	artifactOpt := artifact.Option{
		DisabledAnalyzers: disabledAnalyzers(),
		SkipFiles:         []string{},
		SkipDirs:          []string{},
	}

	builtinPolicyPaths, err := operation.InitBuiltinPolicies(ctx, true)
	if err != nil {
		log.Debugf("trivy policy error: %v\n%v", err, builtinPolicyPaths)
	}

	configScannerOptions := config.ScannerOption{
		Trace:        false,
		Namespaces:   append([]string{}, defaultPolicyNamespace),
		PolicyPaths:  append([]string{}, builtinPolicyPaths...),
		DataPaths:    []string{},
		FilePatterns: []string{},
	}

	artifactArtifact, err := image2.NewArtifact(typesImage, remoteCache, artifactOpt, configScannerOptions)
	if err != nil {
		log.Debugf("trivy artifact error: %v\n%v", err, artifactArtifact)
		cleanup()
	}

	trivyScanner := scanner2.NewScanner(scanner, artifactArtifact)
	if err != nil {
		log.Debugf("trivy scanner error: %v\n%v", err, trivyScanner)
	}
	defer cleanup()

	scanOptions := types.ScanOptions{
		VulnType:            []string{"os", "library"},
		SecurityChecks:      []string{"vuln"},
		ScanRemovedPackages: true,
		ListAllPackages:     true,
	}

	report, err := trivyScanner.ScanArtifact(ctx, scanOptions)
	if err != nil {
		log.Debugf("trivy report error: %v\n%v", err, report)
	}

	return &report, err
}
