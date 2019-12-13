package versioning

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mcuadros/go-version"
)

const (
	validReleaseSemverRegex = "^(v?[0-9]*\\.?[0-9]*\\.?[0-9]*)$"
	//	validSemverRegex        = "^(v?[0-9]*\\.?[0-9]*\\.?[0-9]*)(-[a-z0-9.]+)?$"
	Major    = "major"
	Minor    = "minor"
	Patch    = "patch"
	Same     = "same"
	Notfound = "notfound"

	Unknown = "UNKNOWN"
	Failure = "FAILURE"
)

var regexRelease *regexp.Regexp

//var regex *regexp.Regexp

func init() {
	var err error
	regexRelease, err = regexp.Compile(validReleaseSemverRegex)
	if err != nil {
		log.Fatalf("Could not create regexRelease %v", err)
	}

	/*	regex, err = regexp.Compile(validSemverRegex)
		if err != nil {
			log.Fatalf("Could not create regex %v", err)
		}*/
}

//FindHigestVersionInList finds the higest version in an list of versions or returns NOTFOUND
func FindHigestVersionInList(versions []string) string {
	log.Debugf("FindHigestVersionInList [%v]", versions)
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
	log.Infof("Comparing version [%s] with latest version [%s]", currentVersion, latestVersion)
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
