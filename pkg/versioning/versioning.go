package versioning

import (
	"errors"

	"github.com/blang/semver/v4"
	log "github.com/sirupsen/logrus"
)

// FindHighestSemVer finds the higest SemVer according the spec
// Note that build numbers are ingored and it will takes the last one in the array
func FindHighestSemVer(versions []string) (string, error) {
	var versionSet = false
	var semverVersion semver.Version
	for _, v := range versions {
		version, err := semver.Parse(v)
		if err == nil {
			if version.GTE(semverVersion) {
				semverVersion = version
				versionSet = true
			}
		} else {
			log.WithField("version", v).Warn("Could not parse version")
		}
	}

	if !versionSet {
		return "", errors.New("Could not find any valid SemVer")
	}

	return semverVersion.String(), nil
}
