package registries

// Parts of the code here are comming from github.com/heroku/docker-registry-client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

const (
	// DockerHub is the default name for the DockerHub registry
	DockerHub = "DockerHub"
	// Quay is the default name for the Quay registry
	Quay = "Quay"
	// Gcr is the default name for the Gcr registry
	Gcr = "Gcr"
	// GcrK8s is the default name for the GcrK8s registry
	GcrK8s = "GcrK8s"
	// Zalando is the default name for the Zalando registry
	Zalando = "Zalando"
	// AuthTypeBasic is the basic auth type
	AuthTypeBasic = "basic"
	// AuthTypeToken is the token auth type
	AuthTypeToken = "token"
	// AuthTypeNone is no auth
	AuthTypeNone = "none"
)

// ErrNoMorePages defines that there are no more pages
var ErrNoMorePages = errors.New("no more pages")

type tagsResponse struct {
	Tags []string `json:"tags"`
}

// ImageRegistry contains all the information about the registry
type ImageRegistry struct {
	Name     string `koanf:"name"`
	URL      string `koanf:"url"`
	AuthType string `koanf:"authType"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Default  bool   `koanf:"default"`
}

var cacheToken = ""

// GetLatestVersion fetches the latest version of the docker image from docker registry
func (r ImageRegistry) GetLatestVersion(name string) string {
	log.Infof("Use registry [%s] to find [%s]", r.Name, name)

	//If docker hub and single name (without /) add library/ to it
	if r.Name == DockerHub && !strings.Contains(name, "/") {
		name = "library/" + name
	}

	cacheToken = "" // reset the token

	pathSuffix := fmt.Sprintf("/v2/%s/tags/list", name)
	tags, err := r.fetch(pathSuffix)
	if err != nil {
		log.Errorf("Could not fetch tags for [%s]", name)
		log.Debugf("Could not fetch tags [%v]", err)
		return versioning.Notfound
	}
	return versioning.FindHigestVersionInList(tags)
}

func (r ImageRegistry) fetch(pathSuffix string) ([]string, error) {
	tags := []string{}

	for {
		var response tagsResponse
		var err error
		pathSuffix, err = r.getPaginatedJSON(pathSuffix, &response)
		switch err {
		case ErrNoMorePages:
			log.Debug("No more pages")
			tags = append(tags, response.Tags...)
			return tags, nil
		case nil:
			log.Debug("Fetch next page")
			tags = append(tags, response.Tags...)
			continue
		default:
			log.Debug("Error occured, stop fetching")
			return nil, err
		}
	}
}

func (r ImageRegistry) getPaginatedJSON(pathSuffix string, response interface{}) (string, error) {
	client, req, err := r.getClientAndRequest(pathSuffix)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("Response code was not 200 but [%v]", resp.StatusCode)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	if err != nil {
		return "", err
	}
	return getNextLink(resp)
}

func (r ImageRegistry) getClientAndRequest(pathSuffix string) (*http.Client, *http.Request, error) {
	url := fmt.Sprintf("https://%s%s", r.URL, pathSuffix)
	log.Debugf("try fetching the following [%s]", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	if r.AuthType == AuthTypeBasic {
		req.SetBasicAuth(r.Username, r.Password)
	}

	if r.AuthType == AuthTypeToken && cacheToken == "" {
		log.Debugf("Fetching auth token")
		if err := r.getToken(url); err != nil {
			return nil, nil, err
		}
	}
	if cacheToken != "" {
		log.Debug("Using cached token")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cacheToken))
	}
	return client, req, nil
}

// Matches an RFC 5988 (https://tools.ietf.org/html/rfc5988#section-5)
// Link header. For example,
//
//    <http://r.example.com/v2/_catalog?n=5&last=tag5>; type="application/json"; rel="next"
//
// The URL is _supposed_ to be wrapped by angle brackets `< ... >`,
// but e.g., quay.io does not include them. Similarly, params like
// `rel="next"` may not have quoted values in the wild.
var nextLinkRE = regexp.MustCompile(`^ *<?([^;>]+)>? *(?:;[^;]*)*; *rel="?next"?(?:;.*)?`)

func getNextLink(resp *http.Response) (string, error) {
	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := nextLinkRE.FindStringSubmatch(link)
		if parts != nil {
			return parts[1], nil
		}
	}
	return "", ErrNoMorePages
}

func (r ImageRegistry) getToken(url string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Check if we need to login and find out the token url
	resp, err := client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("Response code was not Unauthorized but [%v]", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Get token url and login to get the token
	tokenURL, err := parsHeaders(resp.Header)
	if err != nil {
		return err
	}

	log.Debugf("Token url [%s]", tokenURL)
	client = &http.Client{}
	req, err = http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return err
	}

	if r.Username != "" || r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Response code was not Oke but [%v]", resp.StatusCode)
	}
	defer resp.Body.Close()

	var authToken authToken
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&authToken)
	if err != nil {
		return err
	}

	cacheToken = authToken.Token
	return nil
}

type authToken struct {
	Token string `json:"token"`
}

// example: Www-Authenticate: Bearer realm="https://auth.docker.io/token",service="r.docker.io",scope="repository:library/ubuntu:pull"
func parsHeaders(headers http.Header) (string, error) {
	authHeader := headers[http.CanonicalHeaderKey("WWW-Authenticate")]
	if len(authHeader) > 1 {
		return "", fmt.Errorf("Not expecting more than one auth header [%v]", authHeader)
	}
	log.Debugf("Auth: [%s]", authHeader[0])
	url := strings.ReplaceAll(authHeader[0], "Bearer realm=", "") // Default registries
	url = strings.ReplaceAll(url, "Basic realm=", "")             //ECR
	log.Debugf("Url: [%s]", url)
	url = strings.Replace(url, ",", "?", 1)
	log.Debugf("Url: [%s]", url)
	url = strings.ReplaceAll(url, ",", "&")
	log.Debugf("Url: [%s]", url)
	url = strings.ReplaceAll(url, "\"", "")
	log.Debugf("Url: [%s]", url)
	return url, nil
}
