package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func container(path, url, name, version string) Image {
	return Image{
		FullPath: path,
		URL:      url,
		Name:     name,
		Version:  version,
	}
}

func TestPodStringToPodStructNoVersion(t *testing.T) {
	pod, _ := ImagePathToImage("test")
	expected := container("test", "docker.io", "library/test", "0")
	assert.Equal(t, expected, pod, "No version")
	assert.Equal(t, "test", pod.WithoutLibrary(), "Remove library")
}

func TestPodStringToPodStructLatestVersion(t *testing.T) {
	pod, _ := ImagePathToImage("test:latest")
	expected := container("test:latest", "docker.io", "library/test", "0")
	assert.Equal(t, expected, pod, "Latest version")
	assert.Equal(t, "test", pod.WithoutLibrary(), "Remove library")
}

func TestPodStringToPodStructOtherVersion(t *testing.T) {
	pod, _ := ImagePathToImage("test:1.3")
	expected := container("test:1.3", "docker.io", "library/test", "1.3")
	assert.Equal(t, expected, pod, "Other version")
	assert.Equal(t, "test", pod.WithoutLibrary(), "Remove library")
}

func TestPodStringToPodStructWithUrl(t *testing.T) {
	pod, _ := ImagePathToImage("gcr.io/test:1.3")
	expected := container("gcr.io/test:1.3", "gcr.io", "test", "1.3")
	assert.Equal(t, expected, pod, "With url")
}

func TestPodStringToPodStructWithImageSubname(t *testing.T) {
	pod, _ := ImagePathToImage("gcr.io/somebody/test:1.3")
	expected := container("gcr.io/somebody/test:1.3", "gcr.io", "somebody/test", "1.3")
	assert.Equal(t, expected, pod, "With image subname")
}

func TestPodStringToPodStructWithPort(t *testing.T) {
	pod, _ := ImagePathToImage("gcr.io:443/somebody/test:1.3")
	expected := container("gcr.io:443/somebody/test:1.3", "gcr.io:443", "somebody/test", "1.3")
	assert.Equal(t, expected, pod, "With port")
}
