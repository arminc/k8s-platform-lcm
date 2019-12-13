package kubernetes

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

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

	if len(namespaces) == 0 {
		log.Debug("No namespaces defined, fetching all")
		namespaces = getAllNamespaces(client)
	} else {
		log.Infof("Get all containers from the namespaces %s", namespaces)
	}

	runningContainers := make(map[string]bool)

	for _, namespace := range namespaces {
		containers := getRunningContainers(client, namespace)
		for key := range containers {
			runningContainers[key] = true
		}
	}

	containers := []Container{}
	for key := range runningContainers {
		container, err := containerStringToContainerStruct(key)
		if err == nil {
			containers = append(containers, container)
		}
	}
	log.Infof("Finished fecthing all containers")
	return containers
}

func containerStringToContainerStruct(containerString string) (Container, error) {
	log.Infof("Extract [%s] to struct", containerString)
	version := "0" // Latest can't be compared
	URL := ""
	fullPath := containerString
	name := containerString

	containerString = strings.Replace(containerString, ":443", "", -1) //Remove 443 if it's there

	if strings.Count(containerString, ":") >= 2 {
		log.Errorf("We do not support urls with ports [%s]", containerString)
		return Container{}, errors.New("We do not support urls with ports")
	}

	if strings.Contains(containerString, ":") {
		//Has a version
		subAndVersion := strings.Split(containerString, ":")
		version = subAndVersion[1]
		containerString = subAndVersion[0]
		name = subAndVersion[0]
	}
	// We assume that image names do not contain a dot
	// When there is a dot it means it has a hostname in front of the image
	if strings.Contains(containerString, ".") {
		urlAndImage := strings.SplitN(containerString, "/", 2)
		URL = urlAndImage[0]
		name = urlAndImage[1]
	}

	//imageAndVersion := strings.Split(urlAndImage[1], ":")
	return Container{
		FullPath: fullPath,
		URL:      URL,
		Name:     name,
		Version:  version,
	}, nil
}

func getKubernetesClient(useLocally bool) *kubernetes.Clientset {
	if useLocally {
		log.Info("Accessing Kubernetes locally")
		kubeconfig := filepath.Join(homeDir(), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Could not find kubernetes config %v", err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("Could not load kubernetes config %v", err)
		}
		return clientset
	}
	log.Info("Accessing Kubernetes inside the cluster")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Could not find kubernetes config in the cluster %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Could not load kubernetes config in the cluster %v", err)
	}

	return clientset
}

func getRunningContainers(client *kubernetes.Clientset, namespace string) map[string]bool {
	containers := make(map[string]bool)
	log.Infof("Fetching containers for [%s] namespace", namespace)
	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Could not fetch pods %v", err)
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			log.Debugf("Container %s", container.Image)
			containers[container.Image] = true
		}
		for _, container := range pod.Spec.InitContainers {
			log.Debugf("InitContainer %s", container.Image)
			containers[container.Image] = true
		}
	}
	return containers
}

func getAllNamespaces(client *kubernetes.Clientset) []string {
	var ns []string
	namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Could not namespaces %v", err)
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
