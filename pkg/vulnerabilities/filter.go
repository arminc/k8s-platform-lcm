package vulnerabilities

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Filter is an interface to filtering vulnerabilities
type Filter interface {
	Vulnerabilities()
}

// FilterData contains all the accepted vulnerabilities that need to be filtered out
type FilterData struct {
	Severities  []string
	Identifiers []Identifier
}

// Identifier holds a matcher to match against and a list of identifiers who are accepted
// In case of an Docker image this can be regular expression like arminc.* or arminc/something.* or explicit arminc/something:1.2.1
// Identifiers need to exactly match the ones passed in
type Identifier struct {
	Match       string
	Identifiers []string
}

// Vulnerabilities filters all accepted vulnerabilities out
// name can be anything, in case of Docker image arminc/something:1.2.1
func (fd *FilterData) Vulnerabilities(name string, vul []Vulnerability) []Vulnerability {
	unnacepted := vul

	if len(fd.Severities) > 0 {
		unnacepted = FilterAcceptedSeverities(unnacepted, fd.Severities)
	}

	if len(fd.Identifiers) > 0 {
		for _, identifier := range fd.Identifiers {
			match, err := regexp.MatchString(identifier.Match, name)
			if err != nil {
				log.WithError(err).Error("Identifier regexp not valid")
			}
			log.Info(match)
			if match {
				unnacepted = FilterAcceptedIdentifiers(unnacepted, identifier.Identifiers)
			}
		}
	}
	return unnacepted
}

// FilterAcceptedSeverities removes accepted severities and returns remaining vulnerabilities
func FilterAcceptedSeverities(vul []Vulnerability, severities []string) []Vulnerability {
	tmp := []Vulnerability{}
	for _, v := range vul {
		for _, severity := range severities {
			if v.Severity == severity {
				tmp = append(tmp, v)
			}
		}
	}
	return difference(vul, tmp)
}

// FilterAcceptedIdentifiers removes accepted identifiers and returns remaining vulnerabilities
func FilterAcceptedIdentifiers(vul []Vulnerability, identifiers []string) []Vulnerability {
	tmp := []Vulnerability{}
	for _, v := range vul {
		for _, identifier := range identifiers {
			if v.Identifier == identifier {
				tmp = append(tmp, v)
			}
		}
	}
	return difference(vul, tmp)
}

// difference returns the diff between two Vulnerability lists (original and accepted ones, returning unaccepted ones)
func difference(a, b []Vulnerability) []Vulnerability {
	target := map[Vulnerability]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []Vulnerability{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}
	return result
}
