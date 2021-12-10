package registries

// Parts of the code here are coming from github.com/heroku/docker-registry-client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"

	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_AWS_REGION = "us-east-1"

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
	// Gitlab is the default name for the Gitlab registry
	Gitlab = "Gitlab"
	// AuthTypeBasic is the basic auth type
	AuthTypeBasic = "basic"
	// AuthTypeToken is the token auth type
	AuthTypeToken = "token"
	// AuthTypeNone is no auth
	AuthTypeNone = "none"
	// AuthTypeECR is ECR auth type
	AuthTypeECR = "ecr"
)

// ErrNoMorePages defines that there are no more pages
var ErrNoMorePages = errors.New("no more pages")

type tagsResponse struct {
	Tags []string `json:"tags"`
}

// ImageRegistry contains all the information about the registry
type ImageRegistry struct {
	Name             string `koanf:"name"`
	URL              string `koanf:"url"`
	AuthType         string `koanf:"authType"`
	Username         string `koanf:"username"`
	Password         string `koanf:"password"`
	Default          bool   `koanf:"default"`
	AllowAllReleases bool
}

// GetLatestVersion fetches the latest version of the docker image from Docker registry
func (r ImageRegistry) GetLatestVersion(name string) string {
	log.WithField("registry", r.Name).WithField("image", name).Debug("Get latest version for Docker image")

	//If docker hub and single name (without /) add library/ to it
	if r.Name == DockerHub && !strings.Contains(name, "/") {
		name = "library/" + name
	}

	cacheToken := "" // reset the token

	pathSuffix := fmt.Sprintf("/v2/%s/tags/list", name)
	tags, err := r.fetch(pathSuffix, cacheToken)
	if err != nil {
		log.WithError(err).WithField("name", name).Error("Could not fetch tags")
		return versioning.Notfound
	}
	return versioning.FindHighestVersionInList(tags, r.AllowAllReleases)
}

func (r ImageRegistry) fetch(pathSuffix string, cacheToken string) ([]string, error) {
	tags := []string{}

	for {
		var response tagsResponse
		var err error
		pathSuffix, err = r.getPaginatedJSON(pathSuffix, &response, cacheToken)
		switch err {
		case ErrNoMorePages:
			tags = append(tags, response.Tags...)
			return tags, nil
		case nil:
			tags = append(tags, response.Tags...)
			continue
		default:
			return nil, err
		}
	}
}

func (r ImageRegistry) getPaginatedJSON(pathSuffix string, response interface{}, cacheToken string) (string, error) {
	client, req, err := r.getClientAndRequest(pathSuffix, cacheToken)
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

func (r ImageRegistry) getClientAndRequest(pathSuffix string, cacheToken string) (*http.Client, *http.Request, error) {
	url := fmt.Sprintf("https://%s%s", r.URL, pathSuffix)
	log.WithField("url", url).Debugf("Try fetching url")
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	if r.AuthType == AuthTypeBasic {
		req.SetBasicAuth(r.Username, r.Password)
	}

	if r.AuthType == AuthTypeToken && cacheToken == "" {
		log.Debug("Need to fetch the auth token")
		cacheToken, err = r.getToken(url)
		if err != nil {
			return nil, nil, err
		}
	}

	if r.AuthType == AuthTypeECR && cacheToken == "" {
		log.Debug("Need to fetch the auth token from AWS")

		config := &aws.Config{Region: aws.String(getAwsRegion())}
		svc := ecr.New(session.New(), config)

		output, err := svc.GetAuthorizationToken(nil)

		if err != nil {
			fmt.Printf("\nError fetching authorization token in region %v\n%v\n", aws.StringValue(config.Region), err.Error())
			os.Exit(1)
		}

		fmt.Printf("\nECR authorization fetched successfully!\n\nAWS Output:\n%v", output)

		cacheToken, err = r.getToken(url)
		if err != nil {
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

func (r ImageRegistry) getToken(url string) (cacheToken string, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Check if we need to login and find out the token url
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("Response code was not Unauthorized but [%v]", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Get token url and login to get the token
	tokenURL, err := parsHeaders(resp.Header)
	if err != nil {
		return "", err
	}

	log.WithField("url", tokenURL).Debug("Token url")
	client = &http.Client{}
	req, err = http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", err
	}

	if r.Username != "" || r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response code was not Oke but [%v]", resp.StatusCode)
	}
	defer resp.Body.Close()

	var authToken authToken
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&authToken)
	if err != nil {
		return "", err
	}

	cacheToken = authToken.Token
	return cacheToken, nil
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
	log.WithField("header", authHeader[0]).Debug("Incoming auth header]")
	url := strings.ReplaceAll(authHeader[0], "Bearer realm=", "") // Default registries
	url = strings.ReplaceAll(url, "Basic realm=", "")             //ECR
	url = strings.Replace(url, ",", "?", 1)
	url = strings.ReplaceAll(url, ",", "&")
	url = strings.ReplaceAll(url, "\"", "")
	return url, nil
}

// Returns the aws region from env var or default region defined in DEFAULT_AWS_REGION constant
func getAwsRegion() string {
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion != "" {
		return awsRegion
	}
	return DEFAULT_AWS_REGION
}
