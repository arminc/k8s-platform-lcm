package internal

import (
	"os"
	"sort"
	"strconv"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	"github.com/arminc/k8s-platform-lcm/internal/registries"
	"github.com/arminc/k8s-platform-lcm/internal/scanning"
	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

// Execute runs all the checks for LCM
func Execute(config config.Config) {
	if config.CommandFlags.DisableKubernetesFetch {
		ExecuteWithoutFetchingContainers(config, []kubernetes.Container{})
	} else {
		containers := kubernetes.GetContainersFromNamespaces(config.Namespaces, config.CommandFlags.LocalKubernetes)
		ExecuteWithoutFetchingContainers(config, containers)
	}
}

// ExecuteWithoutFetchingContainers used for passing in containers without fetching them from Kubernetes
func ExecuteWithoutFetchingContainers(config config.Config, containers []kubernetes.Container) {
	containers = getExtraImages(config.Images, containers)
	info := getLatestVersionsForContainers(containers, config.ImageRegistries)
	info = getVulnerabilities(info, config)
	prettyPrint(info)
	charts := getLatestVersionsForHelmCharts(config.Namespaces, config.CommandFlags.LocalKubernetes)
	prettyPrintChartInfo(charts)
	tools := getLatestVersionsForTools(config.Tools, config.ToolRegistries)
	prettyPrintToolInfo(tools)
}

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

func getLatestVersionsForHelmCharts(namespaces []string, local bool) []ChartInfo {
	var chartInfo []ChartInfo
	charts := kubernetes.GetHelmChartsFromNamespaces(namespaces, local)
	for _, chart := range charts {
		version := registries.GetLatestVersionFromHelm(chart.Name)
		chartInfo = append(chartInfo, ChartInfo{
			Chart:         chart,
			LatestVersion: version,
		})
	}
	return chartInfo
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

func getLatestVersionsForTools(tools []registries.Tool, registries registries.ToolRegistries) []ToolInfo {
	var toolInfo []ToolInfo
	for _, tool := range tools {
		version := registries.GetLatestVersionForTool(tool)
		toolInfo = append(toolInfo, ToolInfo{
			Tool:          tool,
			LatestVersion: version,
		})
	}
	return toolInfo
}

func getVulnerabilities(info []ContainerInfo, config config.Config) []ContainerInfo {
	infoWithVul := []ContainerInfo{}
	for _, ci := range info {
		vulnerabilities := config.ImageScanners.GetVulnerabilities(ci.Container.Name, ci.Container.Version)
		ci.Cves = vulnerabilities
		infoWithVul = append(infoWithVul, ci)
		log.Infof("print me %v", ci)
	}
	return infoWithVul
}

// ContainerInfo contains pod information about the container, its version info and security
type ContainerInfo struct {
	Container     kubernetes.Container
	LatestVersion string
	VersionStatus string
	Fetched       bool
	Cves          []string
}

func getLatestVersionsForContainers(containers []kubernetes.Container, registries registries.ImageRegistries) []ContainerInfo {
	info := []ContainerInfo{}
	for _, container := range containers {
		version := registries.GetLatestVersionForImage(container.Name, container.URL)
		info = append(info, ContainerInfo{
			Container:     container,
			LatestVersion: version,
			VersionStatus: versioning.DetermineLifeCycleStatus(version, container.Version),
		})
	}
	return info
}

func prettyPrint(info []ContainerInfo) {
	sort.Slice(info, func(i, j int) bool {
		return info[i].Container.Name < info[j].Container.Name
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Image", "Version", "Latest", "Cves"})
	table.SetColumnAlignment([]int{3, 1, 1, 3})

	for _, container := range info {
		cve := strconv.Itoa(len(container.Cves))
		if len(container.Cves) == 1 && container.Cves[0] == scanning.ERROR {
			cve = scanning.ERROR
		}

		row := []string{
			container.Container.Name,
			container.Container.Version,
			container.LatestVersion,
			cve,
		}
		table.Append(row)
	}
	table.Render()
}

func prettyPrintToolInfo(tools []ToolInfo) {
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Tool.Repo < tools[j].Tool.Repo
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tool", "Version", "Latest"})
	table.SetColumnAlignment([]int{3, 1, 1})

	for _, tool := range tools {
		row := []string{
			tool.Tool.Repo,
			tool.Tool.Version,
			tool.LatestVersion,
		}
		table.Append(row)
	}
	table.Render()
}

func prettyPrintChartInfo(charts []ChartInfo) {
	sort.Slice(charts, func(i, j int) bool {
		return charts[i].Chart.Name < charts[j].Chart.Name
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Chart", "Version", "Latest"})
	table.SetColumnAlignment([]int{3, 1, 1})

	for _, chart := range charts {
		row := []string{
			chart.Chart.Name,
			chart.Chart.Version,
			chart.LatestVersion,
		}
		table.Append(row)
	}
	table.Render()
}
