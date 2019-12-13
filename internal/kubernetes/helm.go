package kubernetes

import (
	"os"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// Chart is helm chart info
type Chart struct {
	Name    string
	Version string
}

// GetHelmChartsFromNamespaces fetches all charts from the namespaces
func GetHelmChartsFromNamespaces(namespaces []string, useLocally bool) []Chart {

	kubeClient := getKubernetesClient(useLocally)

	if len(namespaces) == 0 {
		log.Debug("No namespaces defined, fetching all")
		namespaces = getAllNamespaces(kubeClient)
	} else {
		log.Infof("Get all containers from the namespaces %s", namespaces)
	}

	var tmpCharts []Chart
	for _, namespace := range namespaces {
		settings := cli.New()
		actionConfig := new(action.Configuration)

		err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Infof)
		if err != nil {
			log.Errorf("Failed to get Helm action config: %v", err)
			continue
		}

		client := action.NewList(actionConfig)
		charts, err := client.Run()
		if err != nil {
			log.Errorf("Failed to run helm command: %v", err)
			continue
		}
		for _, chart := range charts {
			tmpCharts = append(tmpCharts, Chart{
				Name:    chart.Name,
				Version: chart.Chart.Metadata.Version,
			})
		}
	}
	return tmpCharts
}
