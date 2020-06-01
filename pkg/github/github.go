package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// TODO add documentation
// specify proper version fo github.com/stretchr/testify/assert + logrus + oauth2

type RepoVersionGetter interface {
	GetLatestVersion(ctx context.Context, owner, repo string) (string, error)
}

type Credentials struct {
	Username string
	Password string
	Token    string
}

func (c Credentials) isUserNamePasswordSet() bool {
	return (c.Username != "" && c.Password != "")
}
func (c Credentials) isTokenSet() bool {
	return c.Token != ""
}

type githubClient struct {
	client *github.Client
}

func NewGithubClient(ctx context.Context, cred Credentials) RepoVersionGetter {
	if cred.isUserNamePasswordSet() {
		auth := github.BasicAuthTransport{
			Username: cred.Username,
			Password: cred.Password,
		}
		return &githubClient{
			client: github.NewClient(auth.Client()),
		}
	}
	if cred.isTokenSet() {
		auth := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cred.Token},
		)
		return &githubClient{
			github.NewClient(oauth2.NewClient(ctx, auth)),
		}
	}
	return &githubClient{
		github.NewClient(nil),
	}
}

func (gc *githubClient) GetLatestVersion(ctx context.Context, owner, repo string) (string, error) {
	return gc.getLatestReleaseVersion(ctx, owner, repo)
}

// The latest release is the most recent non-prerelease, non-draft release, sorted by the created_at attribute. The created_at attribute is the date of the commit used for the release, and not the date when the release was drafted or published.
func (gc *githubClient) getLatestReleaseVersion(ctx context.Context, owner string, repo string) (string, error) {
	release, response, err := gc.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.WithField("repo", owner+"/"+repo).Errorf("Error fetching latest version: err: %s", err)
		return "", err
	}
	if response.StatusCode != 200 {
		log.WithField("repo", owner+"/"+repo).Errorf("Error fetching latest version: http-status: %s", response.Status)
		return "", fmt.Errorf("Error fetching latest version: %s", response.Status)
	}
	return release.GetTagName(), nil
}
