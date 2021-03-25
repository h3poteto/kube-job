package job

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newClient(configFile string) (*kubernetes.Clientset, error) {
	_, err := os.Stat(configFile)
	if err != nil {
		kubeConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		return kubernetes.NewForConfig(kubeConfig)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}
