package kubernetes

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Container holds the info of the container running in the cluster
type Container struct {
	FullName string
	URL      string
	Name     string
	Version  string
}

// Kube is an interface that wraps calls to Kubernetes cluster
type Kube interface {
	GetAllNamespaces() ([]string, error)
	GetRunningContainers(namespace string) ([]string, error)
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
		k8s, err = GetLocalKubernetesClient()
	} else {
		k8s, err = GetKubernetesClient()
	}
	if err != nil {
		return nil, err
	}
	return &k8sClient{
		client: k8s,
	}, nil
}

// GetContainersFromNamespaces fetches all containers and init containers
func (k k8sClient) GetContainersFromNamespaces(namespaces []string) ([]Container, error) {
	namespaces, err := k.getNamespaces(namespaces)
	if err != nil {
		return nil, err
	}
	log.WithField("namespaces", namespaces).Debug("Using following namespaces")

	runningContainers := make(map[string]bool)

	for _, namespace := range namespaces {
		containers, err := k.GetRunningContainers(namespace)
		if err != nil {
			log.WithField("namespace", namespace).WithError(err).Error("Failed to fetch pods, skip")
		}
		for _, container := range containers {
			runningContainers[container] = true
		}
		log.WithField("namespace", namespace).WithField("containers", containers).Debug("Using following namespaces")
	}

	containers := []Container{}
	for key := range runningContainers {
		container, err := ImageStringToContainerStruct(key)
		if err == nil {
			containers = append(containers, container)
		}
	}
	log.Info("Finished fecthing all containers")
	return containers, nil
}

// getNamespaces returns all namesapces or just the ones that are defined
func (k k8sClient) getNamespaces(namespaces []string) ([]string, error) {
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

// GetRunningContainers returns running and init containers, not deduplicated
func (k k8sClient) GetRunningContainers(namespace string) ([]string, error) {
	containers := []string{}
	log.WithField("namespace", namespace).Debug("Fetching containers for namespace")
	pods, err := k.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not fetch pods")
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
