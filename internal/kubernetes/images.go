package kubernetes

import (
	"github.com/docker/distribution/reference"
	log "github.com/sirupsen/logrus"
)

// ImageStringToContainerStruct converts image string to container information
func ImageStringToContainerStruct(containerString string) (Container, error) {
	image, err := reference.ParseNormalizedNamed(containerString)
	if err != nil {
		log.WithError(err).Error("Failed to pars image name")
		return Container{}, err
	}
	image = reference.TagNameOnly(image) // adds tag latest if no tag is set

	version := image.(reference.NamedTagged).Tag()
	if version == "latest" {
		version = "0" // tag 'latest' can't be compared
	}

	return Container{
		FullPath: containerString,
		URL:      reference.Domain(image),
		Name:     reference.Path(image),
		Version:  version,
	}, nil

}
