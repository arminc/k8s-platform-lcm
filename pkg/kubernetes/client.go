package kubernetes

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetLocalKubernetesClient returns Kubernetes client for local use
func GetLocalKubernetesClient() (*kubernetes.Clientset, error) {
	home := homeDir()
	log.WithField("homedir", home).Debug("Creating Kubernetes client locally")
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "Could not find kubernetes config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not load kubernetes config")
	}
	return clientset, nil
}

// GetKubernetesClient return Kubernetes client for in cluster use
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	log.Debug("Accessing Kubernetes inside the cluster")
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Could not find kubernetes config in the cluster")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not load kubernetes config in the cluster")
	}

	return clientset, nil
}

// homeDir gets the home dir for specific OS
func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE") // windows
}
