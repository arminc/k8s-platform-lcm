package internal

import (
	"os"
	"strconv"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	"github.com/olekukonko/tablewriter"
)

func prettyPrintContainerInfo(info []ContainerInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Image", "Version", "Latest", "Cves"})
	table.SetColumnAlignment([]int{3, 1, 1, 3})

	for _, container := range info {
		row := []string{
			container.Container.Name,
			container.Container.Version,
			container.LatestVersion,
			container.GetCveStatus(),
		}
		table.Append(row)
	}
	table.Render()
}

func prettyPrintGitHubInfo(gitHub []GitHubInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Tool", "Version", "Latest"})
	table.SetColumnAlignment([]int{3, 1, 1})

	for _, info := range gitHub {
		row := []string{
			info.Repo,
			info.Version,
			info.LatestVersion,
		}
		table.Append(row)
	}
	table.Render()
}

func prettyPrintChartInfo(charts []ChartInfo) {
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

// GetCveStatus shows the status based on the cve's
func (c ContainerInfo) GetCveStatus() string {
	cve := strconv.Itoa(len(c.Cves))

	if len(c.Cves) == 1 {
		switch c.Cves[0] {
		case versioning.Failure:
			cve = versioning.Failure
		case versioning.Nodata:
			cve = versioning.Nodata
		}
	}
	return cve
}

// GetStatus shows the status based on version and cve status
func (c ContainerInfo) GetStatus() string {
	if c.LatestVersion == versioning.Notfound {
		return c.LatestVersion
	} else if c.GetCveStatus() == versioning.Failure || c.GetCveStatus() == versioning.Nodata {
		return c.GetCveStatus()
	} else if len(c.Cves) >= 1 {
		return versioning.Failure
	}
	return versioning.DetermineLifeCycleStatus(c.LatestVersion, c.Container.Version)
}

// GetStatus shows status for the chart
func (c ChartInfo) GetStatus() string {
	if c.LatestVersion == versioning.Failure {
		return c.LatestVersion
	}
	return versioning.DetermineLifeCycleStatus(c.LatestVersion, c.Chart.Version)
}

// GetStatus shows the status for the tool
func (t GitHubInfo) GetStatus() string {
	if t.LatestVersion == versioning.Notfound || t.LatestVersion == versioning.Failure {
		return t.LatestVersion
	}
	return versioning.DetermineLifeCycleStatus(t.LatestVersion, t.Version)
}
