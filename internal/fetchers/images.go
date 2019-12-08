package fetchers

// Parts of the code here are comming from github.com/heroku/docker-registry-client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/utils"
	log "github.com/sirupsen/logrus"
)

// ErrNoMorePages defines that there are no more pages
var ErrNoMorePages = errors.New("no more pages")

type tagsResponse struct {
	Tags []string `json:"tags"`
}

var cacheToken = ""

// GetLatestImageVersionFromRegistry fetches the latest versiono of the docker image from docker hub
func GetLatestImageVersionFromRegistry(name string, registry config.ImageRegistry) string {
	//If docker hub and single name (without /) add library/ to it
	if registry.Name == config.DockerHub && !strings.Contains(name, "/") {
		name = "library/" + name
	}

	cacheToken = "" // reset the token

	pathSuffix := fmt.Sprintf("/v2/%s/tags/list", name)
	tags, err := fetch(pathSuffix, registry)
	if err != nil {
		log.Error("Could not fetch tags")
		log.Debugf("Could not fetch tags [%v]", err)
		return utils.Notfound
	}
	return utils.FindHigestVersionInList(tags)
}

func fetch(pathSuffix string, registry config.ImageRegistry) ([]string, error) {
	tags := []string{}

	for {
		var response tagsResponse
		var err error
		pathSuffix, err = getPaginatedJSON(pathSuffix, registry, &response)
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

func getPaginatedJSON(pathSuffix string, registry config.ImageRegistry, response interface{}) (string, error) {
	client, req, err := getClientAndRequest(pathSuffix, registry)
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

func getClientAndRequest(pathSuffix string, registry config.ImageRegistry) (*http.Client, *http.Request, error) {
	url := fmt.Sprintf("https://%s%s", registry.URL, pathSuffix)
	log.Debugf("try fetching the following [%s]", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	if registry.AuthType == config.AuthTypeBasic {
		req.SetBasicAuth(registry.Username, registry.Password)
	}

	if registry.AuthType == config.AuthTypeToken && cacheToken == "" {
		log.Infof("Fetching auth token")
		if err := getToken(url, registry); err != nil {
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
//    <http://registry.example.com/v2/_catalog?n=5&last=tag5>; type="application/json"; rel="next"
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

func getToken(url string, registry config.ImageRegistry) error {
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

	if registry.Username != "" || registry.Password != "" {
		req.SetBasicAuth(registry.Username, registry.Password)
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

// example: Www-Authenticate: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:library/ubuntu:pull"
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
