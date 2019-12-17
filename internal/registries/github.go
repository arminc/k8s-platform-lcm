package registries

import (
	"context"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	"github.com/google/go-github/v28/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// GitHubConfig configuration for GitHub
type GitHubConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Token    string `koanf:"token"`
}

func (g GitHubConfig) isUserNamePasswordSet() bool {
	return (g.Username != "" && g.Password != "")
}
func (g GitHubConfig) isTokenSet() bool {
	return g.Token != ""
}

// GetLatestVersion gets the latest version for a tool from GitHub
func (g GitHubConfig) GetLatestVersion(owner, repo, version string) string {
	ctx := context.Background()
	client := g.getClient(ctx)

	release, response, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			log.WithField("tool", owner+"/"+repo).Error("Hit the rate limit")
			return versioning.Failure
		}
		// If the repository isn't working with releases, just get the latest tag
		return getTags(owner, repo, client)
	}
	if response.StatusCode != 200 {
		log.WithField("tool", owner+"/"+repo).WithField("code", response.StatusCode).Error("Response code was not oke")
		return versioning.Failure
	}
	return release.GetTagName()
}

func getTags(owner string, repo string, client *github.Client) string {
	opt := &github.ListOptions{PerPage: 10}

	var allTags []string
	for {
		tags, resp, err := client.Repositories.ListTags(context.Background(), owner, repo, opt)
		if err != nil {
			log.WithField("tool", owner+"/"+repo).WithError(err).Error("Could not fetch version")
			return versioning.Notfound
		}
		for _, tag := range tags {
			allTags = append(allTags, *tag.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return versioning.FindHighestVersionInList(allTags)
}

func (g GitHubConfig) getClient(ctx context.Context) *github.Client {
	if g.isUserNamePasswordSet() {
		auth := github.BasicAuthTransport{
			Username: g.Username,
			Password: g.Password,
		}
		return github.NewClient(auth.Client())
	} else if g.isTokenSet() {
		auth := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.Token},
		)
		return github.NewClient(oauth2.NewClient(ctx, auth))
	}
	return github.NewClient(nil)
}
