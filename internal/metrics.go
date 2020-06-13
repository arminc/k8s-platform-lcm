package internal

import (
	"context"
	"net/http"

	"github.com/arminc/k8s-platform-lcm/internal/config"
	"github.com/arminc/k8s-platform-lcm/internal/versioning"
	"github.com/arminc/k8s-platform-lcm/pkg/kubernetes"
	log "github.com/sirupsen/logrus"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	prometheusHandler = promhttp.Handler()
	imageStats        = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "image_info",
		Help: "Information on image releases",
	}, []string{
		"image",
		"version",
		"latestVersion",
		"registry",
	})
	chartStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "chart_info",
		Help: "Information on chart releases",
	}, []string{
		"chart",
		"version",
		"latestVersion",
	})
	githubStats = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_info",
		Help: "Information on github tool/application releases",
	}, []string{
		"tool",
		"version",
		"latestVersion",
	})
)

func runStats(config config.Config) {
	ctx := context.Background()

	imageStats.Reset()
	githubStats.Reset()
	chartStats.Reset()

	//charts = getLatestVersionsForHelmCharts(config.HelmRegistries, config.Namespaces, config.RunningLocally(), clients)
	var containers []kubernetes.Image
	var charts []ChartInfo
	if config.IsKubernetesFetchEnabled() {
		kube, err := kubernetes.NewKubeClient(config.RunningLocally())
		if err != nil {
			log.WithError(err).Error("Could not create a kubernetes client")
		} else {
			c, err := kube.GetImagesFromNamespaces(config.Namespaces)
			if err != nil {
				log.WithError(err).Error("Could not fetch image info from kubernetes")
			} else {
				containers = c
			}
		}
		charts = getLatestVersionsForHelmCharts(config.HelmRegistries, config.Namespaces, config.RunningLocally())
	}

	//// charts
	for _, item := range charts {
		chart := item.Chart.Name
		version := item.Chart.Version
		latestVersion := item.LatestVersion
		getHighestVersion := versioning.FindHighestVersionInList([]string{version, latestVersion}, true)
		status := 0.0
		if version == getHighestVersion {
			status = 1.0
		}
		chartStats.WithLabelValues(chart, version, latestVersion).Set(status)
	}

	containers = getExtraImages(config.Images, containers)
	// docker images related
	containers = getExtraImages(config.Images, containers)
	info := getLatestVersionsForContainers(containers, config.ImageRegistries)
	for _, item := range info {
		image := item.Container.Name
		registry := item.Container.URL
		version := item.Container.Version
		latestVersion := item.LatestVersion
		getHighestVersion := versioning.FindHighestVersionInList([]string{version, latestVersion}, true)
		status := 0.0
		if version == getHighestVersion {
			status = 1.0
		}
		imageStats.WithLabelValues(image, version, latestVersion, registry).Set(status)
	}

	// github released versions
	github := getLatestVersionsForGitHub(ctx, config.GitHub)
	for _, item := range github {
		tool := item.Repo
		version := item.Version
		latestVersion := item.LatestVersion
		getHighestVersion := versioning.FindHighestVersionInList([]string{version, latestVersion}, true)
		status := 0.0
		if version == getHighestVersion {
			status = 1.0
		}
		githubStats.WithLabelValues(tool, version, latestVersion).Set(status)
	}
}

// StartMetricsServer starts the server
func StartMetricsServer(config config.Config) {

	http.HandleFunc("/metrics", newStatsHandler(config))
	log.Fatal(http.ListenAndServe(":9572", nil))
}

func newStatsHandler(config config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runStats(config)
		prometheusHandler.ServeHTTP(w, r)
	}
}
