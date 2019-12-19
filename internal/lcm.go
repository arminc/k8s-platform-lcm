package internal

import (
	"sort"
	"time"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	"github.com/arminc/k8s-platform-lcm/internal/registries"
)

// ToolInfo contains tool information with the latest version
type ToolInfo struct {
	Tool          registries.Tool
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

	WebDataVar.Status = "Running"

	var containers = []kubernetes.Container{}
	if config.IsKubernetesFetchEnabled() {
		containers = kubernetes.GetContainersFromNamespaces(config.Namespaces, config.RunningLocally())
	}

	containers = getExtraImages(config.Images, containers)
	info := getLatestVersionsForContainers(containers, config.ImageRegistries)
	info = getVulnerabilities(info, config)
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

	tools := getLatestVersionsForTools(config.Tools, config.ToolRegistries)
	if config.PrettyPrintAllowed() {
		prettyPrintToolInfo(tools)
	}
	WebDataVar.ToolInfo = tools
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
	containerInfo := []ContainerInfo{}
	for _, container := range containers {
		version := registries.GetLatestVersionForImage(container.Name, container.URL)
		containerInfo = append(containerInfo, ContainerInfo{
			Container:     container,
			LatestVersion: version,
		})
	}

	sort.Slice(containerInfo, func(i, j int) bool {
		return containerInfo[i].Container.Name < containerInfo[j].Container.Name
	})
	return containerInfo
}

func getVulnerabilities(containerInfo []ContainerInfo, config config.Config) []ContainerInfo {
	containerInfoWithVul := []ContainerInfo{}
	for _, ci := range containerInfo {
		vulnerabilities := config.ImageScanners.GetVulnerabilities(ci.Container.Name, ci.Container.Version)
		ci.Cves = vulnerabilities
		containerInfoWithVul = append(containerInfoWithVul, ci)
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

func getLatestVersionsForTools(tools []registries.Tool, registries registries.ToolRegistries) []ToolInfo {
	var toolInfo []ToolInfo
	for _, tool := range tools {
		version := registries.GetLatestVersionForTool(tool)
		toolInfo = append(toolInfo, ToolInfo{
			Tool:          tool,
			LatestVersion: version,
		})
	}

	sort.Slice(toolInfo, func(i, j int) bool {
		return toolInfo[i].Tool.Repo < toolInfo[j].Tool.Repo
	})
	return toolInfo
}
