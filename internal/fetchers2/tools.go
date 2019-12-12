package fetchers2

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// ToolRegistries all the tool registries like GitHub
type ToolRegistries struct {
	GitHub GitHubConfig `koanf:"gitHub"`
}

// Tool contains tool that needs to be checked for new version
type Tool struct {
	Repo    string `koanf:"repo"`
	Version string `koanf:"version"`
}

func (t Tool) getRepoAndOwner() (string, string) {
	owner := strings.Split(t.Repo, "/")[0]
	repo := strings.Split(t.Repo, "/")[1]
	return owner, repo
}

// GetLatestVersionForTool gets the latest version for tool
func (t ToolRegistries) GetLatestVersionForTool(tool Tool) string {
	log.Debugf("Finding latest version for [%s]", tool.Repo)
	owner, repo := tool.getRepoAndOwner()
	return t.GitHub.GetLatestVersion(owner, repo, tool.Version)
}
