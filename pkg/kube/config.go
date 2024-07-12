// Package kube provides utilities for working with Kubernetes configurations.
package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeConfigPath returns the path to the Kubernetes configuration file.
// If the KUBECONFIG environment variable is set, it returns the first path
// in the list that matches the current context.
func GetKubeConfigPath() string {
	defaultPath := ""

	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		return defaultPath
	}

	paths := filepath.SplitList(kubeConfig)
	if len(paths) == 1 {
		return paths[0]
	}

	var currentContext string
	for _, path := range paths {
		config, err := clientcmd.LoadFromFile(path)
		if err != nil {
			continue
		}
		if currentContext == "" {
			currentContext = config.CurrentContext
		}
		_, ok := config.Contexts[currentContext]
		if ok {
			return path
		}
	}
	return defaultPath
}
