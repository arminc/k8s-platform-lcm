package internal

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/arminc/k8s-platform-lcm/pkg/github"
	"github.com/arminc/k8s-platform-lcm/pkg/kubernetes"
	"github.com/arminc/k8s-platform-lcm/pkg/trivy"
	"github.com/arminc/k8s-platform-lcm/pkg/vulnerabilities"
	"github.com/arminc/k8s-platform-lcm/pkg/xray"
	log "github.com/sirupsen/logrus"
)

// GitHubInfo contains information with the latest version from GitHub repo
type GitHubInfo struct {
	Repo          string
	Version       string
	LatestVersion string
}

// ChartInfo contains helm chart information with the latest version
type ChartInfo struct {
	Chart         kubernetes.Chart
	LatestVersion string
}

// ContainerInfo contains pod information about the container, its version info, and security
type ContainerInfo struct {
	Container                  kubernetes.Image
	LatestVersion              string
	Fetched                    bool
	Vulnerabilities            []vulnerabilities.Vulnerability
	VulnerabilitiesNotAccepted int
}

// Execute runs all the checks for LCM
func Execute(config config.Config) {
	ctx := context.Background()

	WebDataVar.Status = "Running"

	var containers = []kubernetes.Image{}
	if config.IsKubernetesFetchEnabled() {
		kube, err := kubernetes.NewKubeClient(config.RunningLocally())
		if err != nil {
			log.WithError(err).Error("Could not create a kubernetes client")
		} else {
			c, err := kube.GetImagesFromNamespaces(ctx, config.Namespaces)
			if err != nil {
				log.WithError(err).Error("Could not fetch image info from kubernetes")
			} else {
				containers = c
			}
		}
	}

	containers = getExtraImages(config.Images, containers)
	info := getLatestVersionsForContainers(containers, config.ImageRegistries)
	if len(config.Xray.URL) > 0 || len(config.Trivy.URL) > 0 {
		info = getVulnerabilities(info, config)
	}
	if config.PrettyPrintAllowed() {
		prettyPrintContainerInfo(info)
	}
	WebDataVar.ContainerInfo = info

	if config.IsKubernetesFetchEnabled() {
		charts := getLatestVersionsForHelmCharts(ctx, config.HelmRegistries, config.Namespaces, config.RunningLocally())
		if config.PrettyPrintAllowed() {
			prettyPrintChartInfo(charts)
		}
		WebDataVar.ChartInfo = charts
	}

	github := getLatestVersionsForGitHub(ctx, config.GitHub)
	if config.PrettyPrintAllowed() {
		prettyPrintGitHubInfo(github)
	}

	if config.PrettyPrintAllowed() && config.CliFlags.Vulnerabilities {
		prettyPrintContainerInfoVulnerabilities(info)
	}

	WebDataVar.GitHubInfo = github
	WebDataVar.Status = "Done"
	WebDataVar.LastTimeFetched = time.Now().Format("15:04:05 02-01-2006")
}

func getExtraImages(images []string, containers []kubernetes.Image) []kubernetes.Image {
	for _, image := range images {
		container, err := kubernetes.ImagePathToImage(image)
		if err == nil {
			containers = append(containers, container)
		}
	}
	return containers
}

func getLatestVersionsForContainers(containers []kubernetes.Image, registries registries.ImageRegistries) []ContainerInfo {
	var wg sync.WaitGroup
	var containerInfo []ContainerInfo
	queue := make(chan ContainerInfo, 1)
	wg.Add(len(containers))
	log.WithField("lcm", "getLatestVersionsForContainers").Debugf("all containers slice is %+v", containers)
	for _, container := range containers {
		log.WithField("lcm", "getLatestVersionsForContainers").Debugf("current container is %+v", container)
		go func(container kubernetes.Image) {
			version := registries.GetLatestVersionForImage(container.Name, container.URL)
			newContainerInfo := ContainerInfo{
				Container:     container,
				LatestVersion: version,
			}
			queue <- newContainerInfo
		}(container)
	}

	go func() {
		for t := range queue {
			containerInfo = append(containerInfo, t)
			wg.Done()
		}
	}()

	wg.Wait()
	log.WithField("lcm", "getLatestVersionsForContainers").Debugf("containerInfo slice is %+v", containerInfo)

	sort.Slice(containerInfo, func(i, j int) bool {
		return containerInfo[i].Container.Name < containerInfo[j].Container.Name
	})
	return containerInfo
}

func getVulnerabilities(containerInfo []ContainerInfo, config config.Config) []ContainerInfo {
	filter := vulnerabilities.NewVulnerabilityFilter(config.VulnerabilityFilterData.Severities, config.VulnerabilityFilterData.Identifiers)
	containerInfoWithVul := []ContainerInfo{}

	var err error
	var trivyClient trivy.Scanner
	var xrayClient xray.Scanner
	var scannerIsXray bool

	if len(config.Xray.URL) > 0 && len(config.Trivy.URL) > 0 {
		log.Warn("Cannot use xray and trivy simultaneously")
		return containerInfoWithVul
	}

	if len(config.Xray.URL) > 0 {
		xrayClient, err = xray.NewXray(config.Xray)
		scannerIsXray = true
	} else if len(config.Trivy.URL) > 0 {
		trivyClient, err = trivy.NewTrivy(config.Trivy, config.ImageRegistries)
		scannerIsXray = false
	}

	if err != nil {
		log.WithError(err).Warn("Could not create vuln scanning client")
		for _, ci := range containerInfo {
			ci.Fetched = false
			containerInfoWithVul = append(containerInfoWithVul, ci)
		}
	} else {
		for _, ci := range containerInfo {
			var vulnera []vulnerabilities.Vulnerability
			if scannerIsXray {
				vulnera, err = xrayClient.GetVulnerabilities(ci.Container.Name, ci.Container.Version, config.Xray.Prefixes)
			} else {
				vulnera, err = trivyClient.GetVulnerabilities(ci.Container.FullPath, ci.Container.Name, ci.Container.URL)
			}
			if err != nil {
				log.WithError(err).WithField("image", ci.Container.Name).Warn("Could not fetch vulnerabilities")
				ci.Fetched = false
			} else {
				ci.Fetched = true
				ci.Vulnerabilities = vulnera
				vulnerabilitiesNotAccepted := filter.Vulnerabilities(ci.Container.Name, vulnera)
				ci.VulnerabilitiesNotAccepted = len(vulnerabilitiesNotAccepted)
			}
			containerInfoWithVul = append(containerInfoWithVul, ci)
		}
	}

	sort.Slice(containerInfoWithVul, func(i, j int) bool {
		return containerInfoWithVul[i].Container.Name < containerInfoWithVul[j].Container.Name
	})
	return containerInfoWithVul
}

func getLatestVersionsForHelmCharts(ctx context.Context, helmRegistries registries.HelmRegistries, namespaces []string, local bool) []ChartInfo {
	var chartInfo []ChartInfo
	helm, err := kubernetes.NewHelmClient(local)
	if err != nil {
		log.WithError(err).Error("Failed to create helm client")
	}

	charts, err := helm.GetHelmChartInfoFromNamespaces(ctx, namespaces)
	if err != nil {
		log.WithError(err).Error("Failed to create fetch helm charts")
		return []ChartInfo{}
	}

	for _, chart := range charts {
		version := helmRegistries.GetLatestVersionFromHelm(chart.Chart)
		chartInfo = append(chartInfo, ChartInfo{
			Chart:         chart,
			LatestVersion: version,
		})
	}

	sort.Slice(chartInfo, func(i, j int) bool {
		return chartInfo[i].Chart.Release < chartInfo[j].Chart.Release
	})
	return chartInfo
}

func getLatestVersionsForGitHub(ctx context.Context, gitHubRepos github.Repos) []GitHubInfo {
	gitHub := github.NewRepoVersionGetter(ctx, gitHubRepos.Credentials)
	var gitHubInfo []GitHubInfo
	for _, repo := range gitHubRepos.Repos {
		version, _ := gitHub.GetLatestVersion(ctx, repo)
		gitHubInfo = append(gitHubInfo, GitHubInfo{
			Repo:          repo.Repo,
			Version:       repo.Version,
			LatestVersion: version,
		})
	}

	sort.Slice(gitHubInfo, func(i, j int) bool {
		return gitHubInfo[i].Repo < gitHubInfo[j].Repo
	})
	return gitHubInfo
}
