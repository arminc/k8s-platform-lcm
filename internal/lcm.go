package internal

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/arminc/k8s-platform-lcm/pkg/github"
	"github.com/arminc/k8s-platform-lcm/pkg/xray"
	"github.com/containerd/containerd/log"
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
	Container     kubernetes.Container
	LatestVersion string
	Fetched       bool
	Cves          []string
}

// Execute runs all the checks for LCM
func Execute(config config.Config) {
	ctx := context.Background()

	WebDataVar.Status = "Running"

	var containers = []kubernetes.Container{}
	if config.IsKubernetesFetchEnabled() {
		containers = kubernetes.GetContainersFromNamespaces(config.Namespaces, config.RunningLocally())
		log.L.WithField("lcm", "FetchEabled").Debugf("Containers are %+v", containers)
	}

	containers = getExtraImages(config.Images, containers)
	log.L.WithField("lcm", "InclExtraImages").Debugf("Incl. ExtraContainers are %+v", containers)
	info := getLatestVersionsForContainers(containers, config.ImageRegistries)
	log.L.WithField("lcm", "LatestVersionsForContainers").Debugf("Info after getLatestVersionsForContainers%+v", info)
	if len(config.Xray.URL) > 0 {
		info = getVulnerabilities(info, config)
	}
	if config.PrettyPrintAllowed() {
		prettyPrintContainerInfo(info)
	}
	WebDataVar.ContainerInfo = info

	if config.IsKubernetesFetchEnabled() {
		charts := getLatestVersionsForHelmCharts(config.HelmRegistries, config.Namespaces, config.RunningLocally())
		if config.PrettyPrintAllowed() {
			prettyPrintChartInfo(charts)
		}
		WebDataVar.ChartInfo = charts
	}

	github := getLatestVersionsForGitHub(ctx, config.GitHub)
	if config.PrettyPrintAllowed() {
		prettyPrintGitHubInfo(github)
	}
	WebDataVar.GitHubInfo = github
	WebDataVar.Status = "Done"
	WebDataVar.LastTimeFetched = time.Now().Format("15:04:05 02-01-2006")
}

func getExtraImages(images []string, containers []kubernetes.Container) []kubernetes.Container {
	for _, image := range images {
		container, err := kubernetes.ImageStringToContainerStruct(image)
		if err == nil {
			containers = append(containers, container)
		}
	}
	return containers
}

func getLatestVersionsForContainers(containers []kubernetes.Container, registries registries.ImageRegistries) []ContainerInfo {
	var wg sync.WaitGroup
	var containerInfo []ContainerInfo
	queue := make(chan ContainerInfo, 1)
	wg.Add(len(containers))
	log.L.WithField("lcm", "getLatestVersionsForContainers").Debugf("all containers slice is %+v", containers)
	for _, container := range containers {
		log.L.WithField("lcm", "getLatestVersionsForContainers").Debugf("current container is %+v", container)
		go func(container kubernetes.Container) {
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
	log.L.WithField("lcm", "getLatestVersionsForContainers").Debugf("containerInfo slice is %+v", containerInfo)

	sort.Slice(containerInfo, func(i, j int) bool {
		return containerInfo[i].Container.Name < containerInfo[j].Container.Name
	})
	return containerInfo
}

func getVulnerabilities(containerInfo []ContainerInfo, config config.Config) []ContainerInfo {
	containerInfoWithVul := []ContainerInfo{}
	xray, err := xray.NewXray(config.Xray)
	if err == nil {
		for _, ci := range containerInfo {
			vulnerabilities, err := xray.GetVulnerabilities(ci.Container.Name, ci.Container.Version, config.Xray.Prefixes)
			if err != nil {
				log.L.WithError(err).WithField("image", ci.Container.Name).Warn("Could not fetch vulnerabilities")
			} else {
				var vul []string
				for _, v := range vulnerabilities {
					vul = append(vul, v.Identifier)
				}
				ci.Cves = vul
				containerInfoWithVul = append(containerInfoWithVul, ci)
			}
		}
	} else {
		log.L.WithError(err).Warn("Could not create Xray client")
	}

	sort.Slice(containerInfoWithVul, func(i, j int) bool {
		return containerInfoWithVul[i].Container.Name < containerInfoWithVul[j].Container.Name
	})
	return containerInfoWithVul
}

func getLatestVersionsForHelmCharts(helmRegistries registries.HelmRegistries, namespaces []string, local bool) []ChartInfo {
	var chartInfo []ChartInfo
	charts := kubernetes.GetHelmChartsFromNamespaces(namespaces, local)
	for _, chart := range charts {
		version := helmRegistries.GetLatestVersionFromHelm(chart.Name)
		chartInfo = append(chartInfo, ChartInfo{
			Chart:         chart,
			LatestVersion: version,
		})
	}

	sort.Slice(chartInfo, func(i, j int) bool {
		return chartInfo[i].Chart.Name < chartInfo[j].Chart.Name
	})
	return chartInfo
}

func getLatestVersionsForGitHub(ctx context.Context, gitHubRepos github.Repos) []GitHubInfo {
	var gitHubInfo []GitHubInfo
	for _, repo := range gitHubRepos.Repos {
		gitHub := github.NewRepoVersionGetter(ctx, gitHubRepos.Credentials)
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
