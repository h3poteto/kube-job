package job

import (
	"context"
	"runtime"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func (m mockedPod) Get(context.Context, string, metav1.GetOptions) (*v1.Pod, error) {
	return m.pod, nil
}

func TestParseLabels(t *testing.T) {
	labels := map[string]string{
		"app":     "job",
		"version": "1",
	}
	parsed := parseLabels(labels)
	if parsed != "app=job,version=1" && parsed != "version=1,app=job" {
		t.Errorf("Parsed label does not match: %s", parsed)
	}
}

func TestDiffPods(t *testing.T) {
	pod1 := v1.Pod{}
	pod1.Name = "pod1"
	pod2 := v1.Pod{}
	pod2.Name = "pod2"
	pod3 := v1.Pod{}
	pod3.Name = "pod3"
	currentPodList := []v1.Pod{
		pod1,
		pod2,
	}
	newPodList := []v1.Pod{
		pod1,
		pod2,
		pod3,
	}
	diff := diffPods(currentPodList, newPodList)
	if len(diff) != 1 {
		t.Error("Diff does not match")
	}
	if diff[0].Name != pod3.Name {
		t.Error("Diff does not match")
	}

}

func TestIsPendingPod(t *testing.T) {
	pendingPod := v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
	result := isPendingPod(pendingPod)
	if !result {
		t.Error("pod should be pending")
	}

	runningPod := v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	result = isPendingPod(runningPod)
	if result {
		t.Error("pod should not be pending")
	}
}

func TestWaitToStartPod(t *testing.T) {
	runningPod := v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	runningPod.Namespace = "test"
	runningPod.Name = "pod-test"
	podMock := mockedPod{
		pod: &runningPod,
	}
	coreV1Mock := mockedCoreV1{
		mockedPod: podMock,
	}
	watcher := &Watcher{
		Container: "alipne",
		client: &mockedKubernetes{
			mockedCore: coreV1Mock,
		},
	}
	ctx := context.Background()
	pod, err := watcher.WaitToStartPod(ctx, runningPod)
	if err != nil {
		t.Error(err)
	}
	if pod.Name != runningPod.Name {
		t.Error("pod does not match")
	}
}

func TestWatchDoesNotLeakGoroutinesWhenNoNewPods(t *testing.T) {
	client := fake.NewSimpleClientset()
	watcher := NewWatcher(client, "alpine")
	job := &batchv1.Job{}
	job.Namespace = "default"
	job.Spec.Template.Labels = map[string]string{"app": "leak-test"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = watcher.Watch(job, ctx)
	}()

	// Wait for the watch loop to reach steady state, then sample the goroutine count twice.
	time.Sleep(2 * time.Second)
	before := runtime.NumGoroutine()
	time.Sleep(4 * time.Second)
	after := runtime.NumGoroutine()

	// With the leak, ~4 goroutines accumulate over 4 seconds; allow up to 2 for scheduler/GC jitter.
	if growth := after - before; growth > 2 {
		t.Errorf("goroutines leaked while watching a job with no new pods: %d -> %d (+%d)", before, after, growth)
	}
}
