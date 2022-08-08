package kubeutil

import (
	"errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func GetRestConfig(configFile string) (restConfig *rest.Config, err error) {
	if _, err = os.Stat(configFile); err == nil {
		restConfig, err = clientcmd.BuildConfigFromFlags("", configFile)
		return restConfig, err
	}
	if errors.Is(err, os.ErrNotExist) {
		restConfig, err = rest.InClusterConfig()
	}
	return restConfig, err
}
