package kubernetes

import (
	"context"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// Chart contains Helm chart info
type Chart struct {
	Release   string
	Chart     string
	Namespace string
	Version   string
}

// Helm is an interface that wraps calls to Kubernetes cluster for fetching Helm information
type Helm interface {
	GetHelmChartInfoFromNamespaces(ctx context.Context, namespaces []string) ([]Chart, error)
}

type k8sHelmClient struct {
	kube Kube
}

// NewHelmClient is used to construct access to Kubernetes cluster for Helm
// It returns an implementation of the Helm client represented as the Helm interface
func NewHelmClient(local bool) (Helm, error) {
	kube, err := NewKubeClient(local)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create Kube client for helm usage")
	}

	return &k8sHelmClient{
		kube: kube,
	}, nil
}

// GetHelmChartInfoFromNamespaces fetches all charts from the namespaces
// It uses the provided namespaces or if it the list is empty it fetches it from all namespaces
// It skips the namespaces on error, trying to fetch as much as possible and returning that information
// It returns empty Chart list on other cases
func (h k8sHelmClient) GetHelmChartInfoFromNamespaces(ctx context.Context, namespaces []string) ([]Chart, error) {
	namespaces, err := getNamespaces(ctx, namespaces, h.kube)
	if err != nil {
		return []Chart{}, err
	}

	var charts []Chart
	for _, namespace := range namespaces {
		log.WithField("namespace", namespace).Info("Fetching helm chart info")
		actionConfig, err := initializeHelmActionConfig(namespace)
		if err != nil {
			log.WithField("namespace", namespace).WithError(err).Error("Could not initialize helm for this namespace")
			continue // loop trough the rest of namespaces
		}

		client := action.NewList(actionConfig)
		chartsInNamespace, err := client.Run()
		if err != nil {
			log.WithField("namespace", namespace).WithError(err).Error("Failed to fetch helm info for this namespace")
			continue // loop trough the rest of namespaces
		}

		for _, chart := range chartsInNamespace {
			charts = append(charts, Chart{
				Release:   chart.Name,
				Version:   chart.Chart.Metadata.Version,
				Chart:     chart.Chart.Metadata.Name,
				Namespace: namespace,
			})
		}
	}
	return charts, nil
}

// initializeHelmActionConfig returns initialized action to be used as Helm client
func initializeHelmActionConfig(namespace string) (*action.Configuration, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Infof)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get Helm action config")
	}
	return actionConfig, nil
}
