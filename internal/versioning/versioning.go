package versioning

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mcuadros/go-version"
)

const (
	validReleaseSemverRegex = "^(v?[0-9]*\\.?[0-9]*\\.?[0-9]*)$"
	// Major means a major difference between two versions
	Major = "MAJOR"
	// Minor means a minor difference between two versions
	Minor = "MINOR"
	// Patch means a patch difference between two versions
	Patch = "PATCH"
	// Same means two versions are the same
	Same = "SAME"
	// Unknown means the difference between the two versions is unknown
	Unknown = "UNKNOWN"

	// Notfound means a version could not be found
	Notfound = "NOTFOUND"
	// Failure means something went wrong went finding the version
	Failure = "FAILURE"
)

var regexRelease *regexp.Regexp

func init() {
	var err error
	regexRelease, err = regexp.Compile(validReleaseSemverRegex)
	if err != nil {
		log.Fatalf("Could not create regexRelease [%v]", err)
	}
}

//FindHighestVersionInList finds the highest version in an list of versions or returns NOTFOUND
func FindHighestVersionInList(versions []string) string {
	log.Debugf("FindHighestVersionInList [%v]", versions)
	latestVersion := "0"

	for _, vers := range versions {
		if !strings.Contains(vers, ".") {
			continue
		}
		if regexRelease.MatchString(vers) {
			if version.CompareSimple(version.Normalize(vers), version.Normalize(latestVersion)) == 1 {
				latestVersion = vers
			}
		}
	}

	if latestVersion != "0" {
		return latestVersion
	}
	return Notfound
}

// DetermineLifeCycleStatus compares two versions to determin the status of the difference
func DetermineLifeCycleStatus(latestVersion string, currentVersion string) string {
	log.Debugf("Comparing version [%s] with latest version [%s]", currentVersion, latestVersion)
	latest := strings.Split(version.Normalize(latestVersion), ".")
	curr := strings.Split(version.Normalize(currentVersion), ".")

	if version.Compare(currentVersion, latestVersion, "=") {
		return Same
	}
	if version.Compare(curr[0], latest[0], "<") {
		return Major
	}

	// has minor
	if len(latest) >= 2 && len(curr) >= 2 {
		if version.Compare(curr[1], latest[1], "<") {
			return Minor
		}
	}

	// has patch
	if len(latest) >= 3 && len(curr) >= 3 {
		if version.Compare(curr[2], latest[2], "<") {
			return Patch
		}
	}

	return Unknown
}
