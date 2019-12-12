package internal

import (
	"os"
	"sort"
	"strconv"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/fetchers"
	"github.com/arminc/k8s-platform-lcm/internal/fetchers2"
	"github.com/arminc/k8s-platform-lcm/internal/kubernetes"
	"github.com/arminc/k8s-platform-lcm/internal/scanning"
	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

// Execute runs all the checks for LCM
func Execute() {
	//pods := kubernetes.GetContainersFromNamespaces(config.LcmConfig.Namespaces, config.ConfigFlags.LocalKubernetes)
	//info := getLatestVersionsForContainers(pods)
	//info = getVulnerabilities(info)
	//prettyPrint(info)
	tools := getLatestVerfionForTools(config.LcmConfig.Tools)
	prettyPrintToolInfo(tools)
}

func FakeExecute(containers []kubernetes.Container) {
	info := getLatestVersionsForContainers(containers)
	info = getVulnerabilities(info)
	prettyPrint(info)
}

type ToolInfo struct {
	Tool          fetchers2.Tool
	LatestVersion string
}

func getLatestVerfionForTools(tools []fetchers2.Tool) []ToolInfo {
	var toolInfo []ToolInfo
	registries := config.LcmConfig.ToolRegistries
	for _, tool := range tools {
		version := registries.GetLatestVersionForTool(tool)
		toolInfo = append(toolInfo, ToolInfo{
			Tool:          tool,
			LatestVersion: version,
		})
	}
	return toolInfo
}

func getVulnerabilities(info []ContainerInfo) []ContainerInfo {
	infoWithVul := []ContainerInfo{}
	for _, ci := range info {
		vulnerabilities := scanning.GetVulnerabilities(ci.Container)
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

func getLatestVersionsForContainers(containers []kubernetes.Container) []ContainerInfo {
	info := []ContainerInfo{}
	for _, container := range containers {
		registry := determinRegistry(container)
		version := fetchers.GetLatestImageVersionFromRegistry(container.Name, registry)
		info = append(info, ContainerInfo{
			Container:     container,
			LatestVersion: version,
			VersionStatus: versioning.DetermineLifeCycleStatus(version, container.Version),
		})
	}
	return info
}

func determinRegistry(container kubernetes.Container) config.ImageRegistry {
	registry, exists := config.LcmConfig.FindRegistryByOverrideByImage(container.Name)
	if exists {
		return registry
	}

	registry, exists = config.LcmConfig.FindRegistryByOverrideByURL(container.URL)
	if exists {
		return registry
	}

	registry, exists = config.LcmConfig.GetDefaultRegistry()
	if exists {
		return registry
	}

	return config.LcmConfig.FindRegistryByURL(container.URL)
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
