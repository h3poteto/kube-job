package job

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// Watcher has client of kubernetes and target container information.
type Watcher struct {
	client *kubernetes.Clientset

	// Target container name.
	Container string
}

// NewWatcher returns a new Watcher struct.
func NewWatcher(client *kubernetes.Clientset, container string) *Watcher {
	return &Watcher{
		client,
		container,
	}
}

// Watch gets pods and tail the logs.
// We must create endless loop because sometimes jobs are configured restartPolicy.
// When restartPolicy is Never, the Job create a new Pod if the specified command is failed.
// So we must trace all Pods even though the Pod is failed.
// And it isn't necessary to stop the loop because the Job is watched in WaitJobComplete.
func (w *Watcher) Watch(job *v1.Job, ctx context.Context) error {
	currentPodList := []corev1.Pod{}
retry:
	for {
		newPodList, err := w.FindPods(job)
		if err != nil {
			return err
		}

		incrementalPodList := diffPods(currentPodList, newPodList)
		go w.WatchPods(ctx, incrementalPodList)

		time.Sleep(1 * time.Second)
		currentPodList = newPodList
		continue retry
	}
}

// WatchPods gets wait to start pod and tail the logs.
func (w *Watcher) WatchPods(ctx context.Context, pods []corev1.Pod) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(pods))

	for _, pod := range pods {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pod, err := w.WaitToStartPod(pod)
			if err != nil {
				errCh <- err
				return
			}
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
			err = readStreamLog(request, pod)
			errCh <- err
		}()
	}

	select {
	case err := <-errCh:
		if err != nil {
			log.Error(err)
			return err
		}
	}
	wg.Wait()
	return nil
}

// FindPods finds pods in
func (w *Watcher) FindPods(job *v1.Job) ([]corev1.Pod, error) {
	labels := parseLabels(job.Spec.Template.Labels)
	listOptions := metav1.ListOptions{
		LabelSelector: labels,
	}
	podList, err := w.client.CoreV1().Pods(job.Namespace).List(listOptions)
	if err != nil {
		return []corev1.Pod{}, err
	}
	return podList.Items, err
}

// WaitToStartPod wait until starting the pod.
// Because the job does not start immediately after call kubernetes API.
// So we have to wait to start the pod, before watch logs.
func (w *Watcher) WaitToStartPod(pod corev1.Pod) (corev1.Pod, error) {
retry:
	for {
		targetPod, err := w.client.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return pod, err
		}

		if !isPendingPod(*targetPod) {
			return *targetPod, nil
		}
		time.Sleep(1 * time.Second)
		continue retry
	}
}

// isPendingPod check the pods whether it have pending container.
func isPendingPod(pod corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodPending {
		return true
	}
	return false
}

// parseLabels parses label sets, and build query string.
func parseLabels(labels map[string]string) string {
	query := []string{}
	for k, v := range labels {
		query = append(query, k+"="+v)
	}
	return strings.Join(query, ",")
}

// readStreamLog reads rest request, and output the log to stdout with stream.
func readStreamLog(request *restclient.Request, pod corev1.Pod) error {
	readCloser, err := request.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()
	_, err = io.Copy(os.Stdout, readCloser)
	return err
}

// diffPods returns diff between the two pods list.
// It returns newPodList - currentPodList, which is incremental Pods list.
func diffPods(currentPodList, newPodList []corev1.Pod) []corev1.Pod {
	var diff []corev1.Pod

	for _, newPod := range newPodList {
		found := false
		for _, currentPod := range currentPodList {
			if currentPod.Name == newPod.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, newPod)
		}
	}
	return diff
}
