package github

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v31/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// RepoVersionGetter is an interface that wrapps calls to GitHub
type RepoVersionGetter interface {
	GetLatestVersion(ctx context.Context, owner, repo string) (string, error)
}

// Credentials represents different credential options that can be used when calling GitHub
// Username/Password is an combination, Token is standalone and is prefered
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

// NewGithubClient returns an implementation of the GitHub client represented as the RepoVersionGetter interface
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
