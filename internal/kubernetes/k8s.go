package kubernetes

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Container holds the info of the container running in the cluster
type Container struct {
	FullPath string
	URL      string
	Name     string
	Version  string
}

// GetContainersFromNamespaces fetches all containers and init containers
func GetContainersFromNamespaces(namespaces []string, useLocally bool) []Container {
	client := getKubernetesClient(useLocally)
	namespaces = getNamespaces(namespaces, client)
	runningContainers := make(map[string]bool)

	for _, namespace := range namespaces {
		containers := getRunningContainers(client, namespace)
		for key := range containers {
			runningContainers[key] = true
		}
	}

	containers := []Container{}
	for key := range runningContainers {
		container, err := ImageStringToContainerStruct(key)
		if err == nil {
			containers = append(containers, container)
		}
	}
	log.Info("Finished fecthing all containers")
	return containers
}

func getKubernetesClient(useLocally bool) *kubernetes.Clientset {
	if useLocally {
		log.Debug("Accessing Kubernetes locally")
		kubeconfig := filepath.Join(homeDir(), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.WithError(err).Fatal("Could not find kubernetes config")
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.WithError(err).Fatal("Could not load kubernetes config")
		}
		return clientset
	}

	log.Debug("Accessing Kubernetes inside the cluster")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Fatal("Could not find kubernetes config in the cluster")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Fatal("Could not load kubernetes config in the cluster")
	}

	return clientset
}

func getRunningContainers(client *kubernetes.Clientset, namespace string) map[string]bool {
	containers := make(map[string]bool)
	log.WithField("namespace", namespace).Info("Fetching containers for namespace")
	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Could not fetch pods")
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			containers[container.Image] = true
		}
		for _, container := range pod.Spec.InitContainers {
			containers[container.Image] = true
		}
	}
	log.WithField("namespace", namespace).WithField("images", containers).Debug("Fetched containers in namespace")
	return containers
}

func getNamespaces(namespaces []string, client *kubernetes.Clientset) []string {
	if len(namespaces) == 0 {
		log.Debug("No namespaces defined, fetching all namespaces from Kubernetes")
		return getAllNamespaces(client)
	}
	log.WithField("namespaces", namespaces).Info("Get all containers from the namespaces")
	return namespaces
}

func getAllNamespaces(client *kubernetes.Clientset) []string {
	var ns []string
	namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Could not fetch namespaces")
	}

	for _, namespace := range namespaces.Items {
		ns = append(ns, namespace.GetObjectMeta().GetName())
	}
	return ns
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE") // windows
}
