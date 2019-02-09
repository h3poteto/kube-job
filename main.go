package main

import (
	"fmt"
	"log"
	"os"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	client, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	pods, err := client.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}

func newClient() (kubernetes.Interface, error) {
	configFile := os.Getenv("KUBECONFIG")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", configFile)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}
