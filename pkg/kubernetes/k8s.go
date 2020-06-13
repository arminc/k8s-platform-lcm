package kubernetes

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Kube is an interface that wraps calls to Kubernetes cluster
type Kube interface {
	GetImagesFromNamespaces(namespaces []string) ([]Image, error)
	GetAllNamespaces() ([]string, error)
	GetImagePathsFromPods(namespace string) ([]string, error)
}
type k8sClient struct {
	client *kubernetes.Clientset
}

// NewKubeClient is used to construct access to Kubernetes cluster
// It returns an implementation of the Kubernetes client represented as the Kube interface
func NewKubeClient(local bool) (Kube, error) {
	var k8s *kubernetes.Clientset
	var err error
	if local {
		k8s, err = CreateLocalKubernetesClient()
	} else {
		k8s, err = CreateKubernetesClient()
	}
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		client: k8s,
	}, nil
}

// GetImagesFromNamespaces fetches all containers and init containers
// It uses the provided namespaces or if it the list is empty it fetches it from all namespaces
// It skips the namespaces on error, trying to fetch as much as possible and returning that information
// It returns empty Image list on other cases
func (k k8sClient) GetImagesFromNamespaces(namespaces []string) ([]Image, error) {
	namespaces, err := getNamespaces(namespaces, k)
	if err != nil {
		return []Image{}, err
	}
	log.WithField("namespaces", namespaces).Debug("Using following namespaces")

	runningContainers := make(map[string]bool)

	for _, namespace := range namespaces {
		containers, err := k.GetImagePathsFromPods(namespace)
		if err != nil {
			log.WithField("namespace", namespace).WithError(err).Error("Failed to fetch pods, skip")
		}
		for _, container := range containers {
			runningContainers[container] = true
		}
		log.WithField("namespace", namespace).WithField("containers", containers).Debug("Using following namespaces")
	}

	containers := []Image{}
	for key := range runningContainers {
		container, err := ImagePathToImage(key)
		if err == nil {
			containers = append(containers, container)
		}
	}
	log.Info("Finished fecthing all containers")
	return containers, nil
}

// getNamespaces returns all namesapces or just the ones that are defined
func getNamespaces(namespaces []string, k Kube) ([]string, error) {
	if len(namespaces) == 0 {
		log.Debug("No namespaces defined, fetching all namespaces from Kubernetes")
		return k.GetAllNamespaces()
	}
	return namespaces, nil
}

// GetAllNamespaces returns all namespaces from the cluster
func (k k8sClient) GetAllNamespaces() ([]string, error) {
	log.Debug("Fetching all namespaces from Kubernetes")
	namespaces, err := k.client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch namespaces")
	}

	var ns []string
	for _, namespace := range namespaces.Items {
		ns = append(ns, namespace.GetObjectMeta().GetName())
	}
	return ns, nil
}

// GetImagePathsFromPods returns Docker image paths from running and init containers in the Pods
// It dos not deduplicated
// It returns empty list on error
func (k k8sClient) GetImagePathsFromPods(namespace string) ([]string, error) {
	containers := []string{}
	log.WithField("namespace", namespace).Debug("Fetching containers for namespace")
	pods, err := k.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return containers, errors.Wrap(err, "Could not fetch pods")
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Image)
		}
		for _, container := range pod.Spec.InitContainers {
			containers = append(containers, container.Image)
		}
	}
	return containers, nil
}
