package job

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type Watcher struct {
	client    *kubernetes.Clientset
	Container string
}

func NewWatcher(client *kubernetes.Clientset, container string) *Watcher {
	return &Watcher{
		client,
		container,
	}
}

// Watch gets pods and tail the logs.
// At first, finds pods from the job definition, and waits to start the pods.
// Next, gets log requests and get the output with stream.
func (w *Watcher) Watch(job *v1.Job, ctx context.Context) error {
	podList, err := w.WaitToStartPods(job)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(podList.Items))
	for _, pod := range podList.Items {
		// Ref: https://github.com/kubernetes/client-go/blob/03bfb9bdcfe5482795b999f39ca3ed9ad42ce5bb/kubernetes/typed/core/v1/pod_expansion.go
		logOptions := corev1.PodLogOptions{
			Container: w.Container,
			Follow:    true,
		}
		// Ref: https://stackoverflow.com/questions/32983228/kubernetes-go-client-api-for-log-of-a-particular-pod
		request := w.client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &logOptions).Context(ctx).
			Param("follow", strconv.FormatBool(true)).
			Param("container", w.Container).
			Param("timestamps", strconv.FormatBool(false))
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := readStreamLog(request, pod)
			errCh <- err
		}()
	}
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	}
	wg.Wait()
	return nil
}

func (w *Watcher) WaitToStartPods(job *v1.Job) (*corev1.PodList, error) {
retry:
	for {
		labels := parseLabels(job.Spec.Template.Labels)
		listOptions := metav1.ListOptions{
			LabelSelector: labels,
		}
		podList, err := w.client.CoreV1().Pods(job.Namespace).List(listOptions)
		if err != nil {
			return nil, err
		}
		if !hasPendingContainer(podList.Items) {
			return podList, nil
		}
		time.Sleep(1 * time.Second)
		continue retry
	}
}

func hasPendingContainer(pods []corev1.Pod) bool {
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodPending {
			return true
		}
	}
	return false
}

func parseLabels(labels map[string]string) string {
	query := []string{}
	for k, v := range labels {
		query = append(query, k+"="+v)
	}
	return strings.Join(query, ",")
}

func readStreamLog(request *restclient.Request, pod corev1.Pod) error {
	readCloser, err := request.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()
	_, err = io.Copy(os.Stdout, readCloser)
	return err
}
