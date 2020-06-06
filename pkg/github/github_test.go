package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateNewTokenClient(t *testing.T) {
	NewGithubClient(context.Background(), "")
}

func TestCreateNewUserPassClient(t *testing.T) {
	NewGithubClientUserPass(context.Background(), "", "")
}

func TestGetReleaseNonExistingRepo(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	_, err := client.GetLatestVersion(context.TODO(), "arminc", "unknown")
	assert.Error(t, err)
}

func TestGetReleaseNonExistend(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	_, err := client.GetLatestVersion(context.TODO(), "arminc", "lcm_empty")
	assert.Error(t, err)
}

func TestGetRelease(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	version, err := client.GetLatestVersion(context.TODO(), "arminc", "lcm_release")
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0", version, "Version should be the same")
}

func TestGetTagsNonExistingRepo(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	_, err := client.GetLatestVersionFromTag(context.TODO(), "arminc", "unknown")
	assert.Error(t, err)
}

func TestGetTagsNonExistend(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	_, err := client.GetLatestVersionFromTag(context.TODO(), "arminc", "lcm_empty")
	assert.Error(t, err)
}

func TestGetTags(t *testing.T) {
	client := NewGithubClientAnonymous(context.Background())
	version, err := client.GetLatestVersionFromTag(context.TODO(), "arminc", "lcm_release")
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0", version, "Version should be the same")
}
