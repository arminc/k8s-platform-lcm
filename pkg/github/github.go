// Package github is used to access GitHub to find latest version in repositories
package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arminc/k8s-platform-lcm/pkg/versioning"
	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// RepoVersionGetter is an interface that wrapps calls to GitHub
type RepoVersionGetter interface {
	GetLatestVersion(ctx context.Context, owner, repo string) (string, error)
	GetLatestVersionFromTag(ctx context.Context, owner, repo string) (string, error)
}

type githubClient struct {
	client *github.Client
}

// NewGithubClient is used for anonymous access to Github
// It returns an implementation of the GitHub client represented as the RepoVersionGetter interface
func NewGithubClientAnonymous(ctx context.Context) RepoVersionGetter {
	return &githubClient{
		github.NewClient(nil),
	}
}

// NewGithubClientUserPass is used for access to GitHub with username and password combination
// It returns an implementation of the GitHub client represented as the RepoVersionGetter interface
func NewGithubClientUserPass(ctx context.Context, username, password string) RepoVersionGetter {
	auth := github.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	return &githubClient{
		client: github.NewClient(auth.Client()),
	}
}

// NewGithubClient is the default way to access GitHub by using a token
// It returns an implementation of the GitHub client represented as the RepoVersionGetter interface
func NewGithubClient(ctx context.Context, token string) RepoVersionGetter {
	auth := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &githubClient{
		github.NewClient(oauth2.NewClient(ctx, auth)),
	}
}

// GetLatestVersion returns the latest release version from GitHub
// The latest release is the most recent non-prerelease, non-draft release, sorted by the created_at attribute. The created_at attribute is the date of the commit used for the release, and not the date when the release was drafted or published.
func (gc *githubClient) GetLatestVersion(ctx context.Context, owner, repo string) (string, error) {
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
// For now this works by comparing all tags to vind the latest, which means tags need to follow semver
func (gc *githubClient) GetLatestVersionFromTag(ctx context.Context, owner, repo string) (string, error) {
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
