package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient returns the client if its possible in cluster, otherwise tries to read HOME
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := GetKubernetesConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetKubernetesConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	} else if err != rest.ErrNotInCluster {
		return nil, err
	}

	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}
