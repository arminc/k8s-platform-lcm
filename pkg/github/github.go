// Package github is used to access GitHub to find latest version in repositories
package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arminc/k8s-platform-lcm/pkg/versioning"
	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Repos contains all the GitHub repositories and security information to watch for new versions on GitHub
type Repos struct {
	Credentials Credentials `koanf:"credentials"`
	Repos       []Repo      `koanf:"repos"`
}

// Credentials contains access details for GitHub
type Credentials struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Token    string `koanf:"token"`
}

// Repo contains repositories that need to be checked for a new version
type Repo struct {
	Repo    string `koanf:"repo"`
	Version string `koanf:"version"`
	UseTag  bool   `koanf:"useTag"`
}

// RepoVersionGetter is an interface that wraps calls to GitHub
type RepoVersionGetter interface {
	GetLatestVersion(ctx context.Context, gitRepo Repo) (string, error)
	GetLatestVersionFromRelease(ctx context.Context, owner string, repo string) (string, error)
	GetLatestVersionFromTag(ctx context.Context, owner string, repo string) (string, error)
}

type githubClient struct {
	client *github.Client
}

// NewRepoVersionGetter is used to construct authenticated or unauthenticated access to GitHub
// It returns an implementation of the GitHub client represented as the RepoVersionGetter interface
func NewRepoVersionGetter(ctx context.Context, credentials Credentials) RepoVersionGetter {
	if credentials.Username != "" && credentials.Password != "" {
		auth := github.BasicAuthTransport{
			Username: credentials.Username,
			Password: credentials.Password,
		}
		return &githubClient{
			client: github.NewClient(auth.Client()),
		}
	}

	if credentials.Token != "" {
		auth := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: credentials.Token},
		)
		return &githubClient{
			github.NewClient(oauth2.NewClient(ctx, auth)),
		}
	}

	return &githubClient{
		github.NewClient(nil),
	}
}

// getRepoAndOwner splits repo "owner/repo" to owner and repo
func (r Repo) getRepoAndOwner() (string, string) {
	owner := strings.Split(r.Repo, "/")[0]
	repo := strings.Split(r.Repo, "/")[1]
	return owner, repo
}

// GetLatestVersion returns latest version from release or tag depending on the setting
func (gc *githubClient) GetLatestVersion(ctx context.Context, gitRepo Repo) (string, error) {
	owner, repo := gitRepo.getRepoAndOwner()
	if gitRepo.UseTag {
		return gc.GetLatestVersionFromTag(ctx, owner, repo)
	}
	return gc.GetLatestVersionFromRelease(ctx, owner, repo)
}

// GetLatestVersionFromRelease returns the latest release version from GitHub
// The latest release is the most recent non-prerelease, non-draft release, sorted by the created_at attribute. The created_at attribute is the date of the commit used for the release, and not the date when the release was drafted or published.
func (gc *githubClient) GetLatestVersionFromRelease(ctx context.Context, owner string, repo string) (string, error) {
	var err error
	version := ""

	err = retryWhenRateLimited(func() error {
		version, err = gc.getLatestReleaseVersion(ctx, owner, repo)
		return err
	})

	return version, err
}

// Actual call to GitHub
func (gc *githubClient) getLatestReleaseVersion(ctx context.Context, owner string, repo string) (string, error) {
	release, response, err := gc.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.WithField("repo", owner+"/"+repo).Warnf("Error fetching latest version: err: %s", err)
		return "", err
	}
	if response.StatusCode != 200 {
		log.WithField("repo", owner+"/"+repo).Warnf("Error fetching latest version: http-status: %s", response.Status)
		return "", fmt.Errorf("Error fetching latest version: %s", response.Status)
	}
	return release.GetTagName(), nil
}

// GetLatestVersionFromTag returns the latest version from GitHub using tags
// For now this works by comparing all tags to find the latest, which means tags need to follow semver
func (gc *githubClient) GetLatestVersionFromTag(ctx context.Context, owner string, repo string) (string, error) {
	var err error
	version := ""

	err = retryWhenRateLimited(func() error {
		version, err = gc.getLatestTag(ctx, owner, repo)
		return err
	})

	return version, err
}

// Actual call to GitHub
func (gc *githubClient) getLatestTag(ctx context.Context, owner string, repo string) (string, error) {
	opt := &github.ListOptions{PerPage: 10}
	var allTags []string
	for {
		tags, response, err := gc.client.Repositories.ListTags(ctx, owner, repo, opt)
		if err != nil {
			log.WithField("repo", owner+"/"+repo).Warnf("Error fetching latest tags: err: %s", err)
			return "", err
		}
		if response.StatusCode != 200 {
			log.WithField("repo", owner+"/"+repo).Warnf("Error fetching latest tags: http-status: %s", response.Status)
			return "", fmt.Errorf("Error fetching latest tag: %s", response.Status)
		}
		for _, tag := range tags {
			allTags = append(allTags, *tag.Name)
		}
		if response.NextPage == 0 {
			break
		}
		opt.Page = response.NextPage
	}
	return versioning.FindHighestSemVer(allTags)
}

// Retry for 5 times if there is an RateLimit error
func retryWhenRateLimited(cb func() error) error {
	retries := 0
	for {
		if retries > 5 {
			return errors.New("To many retries, stopping")
		}
		retries++

		err := cb()
		if err != nil {
			rerr, ok := err.(*github.RateLimitError)
			if ok {
				var d = time.Until(rerr.Rate.Reset.Time)
				log.Warnf("hit rate limit, sleeping for %.0f min", d.Minutes())
				time.Sleep(d)
				continue
			}
			aerr, ok := err.(*github.AbuseRateLimitError)
			if ok {
				var d = aerr.GetRetryAfter()
				log.Warnf("hit abuse mechanism, sleeping for %.f min", d.Minutes())
				time.Sleep(d)
				continue
			}
			log.Warnf("Error calling github web-api: %s", err)
			return err
		}
		return nil
	}
}
