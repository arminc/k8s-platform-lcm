package trivy

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/arminc/k8s-platform-lcm/internal/registries"

	"github.com/aquasecurity/fanal/analyzer"
	"github.com/aquasecurity/fanal/analyzer/config"
	"github.com/aquasecurity/fanal/artifact"
	image2 "github.com/aquasecurity/fanal/artifact/image"
	"github.com/aquasecurity/fanal/image"
	types2 "github.com/aquasecurity/fanal/types"
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
	analyzers = append(analyzers, analyzer.TypeApkCommand)

	return analyzers
}

func NewClient(remoteUrl string) (s *client.Scanner, err error) {
	log.Debug("Getting trivy client")

	url := client.RemoteURL(remoteUrl)
	customHeaders := client.CustomHeaders(make(map[string][]string))

	scannerScanner := client.NewProtobufClient(url)
	clientScanner := client.NewScanner(customHeaders, scannerScanner)

	return &clientScanner, nil
}

func getDockerOptions(registry registries.ImageRegistry) (o *types2.DockerOption, err error) {
	switch registry.AuthType {
	case "none":
		return &types2.DockerOption{}, nil
	case "basic":
		return &types2.DockerOption{
			UserName: registry.Username,
			Password: registry.Password,
		}, nil
	case "token":
		return &types2.DockerOption{}, nil
	case "ecr":
		// get credentials for ECR
		roleProvider := &ec2rolecreds.EC2RoleProvider{
			Client: ec2metadata.New(session.New()),
		}

		creds := credentials.NewCredentials(roleProvider)
		credVal, err := creds.Get()
		if err != nil {
			return nil, err
		}

		return &types2.DockerOption{
			AwsAccessKey:    credVal.AccessKeyID,
			AwsSecretKey:    credVal.SecretAccessKey,
			AwsSessionToken: credVal.SessionToken,
			AwsRegion:       registry.Region,
		}, nil
	default:
		return &types2.DockerOption{}, nil
	}
}

func Run(scanner *client.Scanner, url string, imageName string, registry registries.ImageRegistry) (r *pkgReport.Report, err error) {
	ctx := context.Background()

	customCacheHeaders := http.Header(make(map[string][]string))

	remoteCache := cache.NewRemoteCache(cache.RemoteURL(url), customCacheHeaders)

	dockerOptions, err := getDockerOptions(registry)
	if err != nil {
		log.Debugf("trivy dockerOptions error: %v", err)
		return nil, err
	}

	typesImage, cleanup, err := image.NewDockerImage(ctx, imageName, *dockerOptions)
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
