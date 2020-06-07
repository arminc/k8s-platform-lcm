package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createRepo(repo string, tag bool) Repo {
	return Repo{
		Repo:    repo,
		Version: "",
		UseTag:  tag,
	}
}

func TestGetReleaseNonExistingRepo(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	_, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/unknown", false))
	assert.Error(t, err)
}

func TestGetReleaseNonExisted(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	_, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/lcm_empty", false))
	assert.Error(t, err)
}

func TestGetRelease(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	version, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/lcm_release", false))
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0", version, "Version should be the same")
}

func TestGetTagsNonExistingRepo(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	_, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/unknown", true))
	assert.Error(t, err)
}

func TestGetTagsNonExisted(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	_, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/lcm_empty", true))
	assert.Error(t, err)
}

func TestGetTags(t *testing.T) {
	client := NewRepoVersionGetter(context.TODO(), Credentials{})
	version, err := client.GetLatestVersion(context.TODO(), createRepo("arminc/lcm_release", true))
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0", version, "Version should be the same")
}
