package kubernetes

import (
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

// Image holds the Docker image information of the container running in the cluster
type Image struct {
	FullPath string
	URL      string
	Name     string
	Version  string
}

// WithoutLibrary returns the Image name without Docker prefix 'library' when using single name Docker repo like ubuntu, etc..
func (c Image) WithoutLibrary() string {
	return strings.Replace(c.Name, "library/", "", 1)
}

// ImagePathToImage converts image string to container information
// In case of an error it returns an empty Image
func ImagePathToImage(imagePath string) (Image, error) {
	image, err := reference.ParseNormalizedNamed(imagePath)
	if err != nil {
		return Image{}, errors.Wrap(err, "Failed to pars image name")
	}
	image = reference.TagNameOnly(image) // adds tag latest if no tag is set

	version := image.(reference.NamedTagged).Tag()
	if version == "latest" {
		version = "0" // tag 'latest' can't be compared
	}

	return Image{
		FullPath: imagePath,
		URL:      reference.Domain(image),
		Name:     reference.Path(image),
		Version:  version,
	}, nil

}
