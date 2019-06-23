package job

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

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
	pod1 := corev1.Pod{}
	pod1.Name = "pod1"
	pod2 := corev1.Pod{}
	pod2.Name = "pod2"
	pod3 := corev1.Pod{}
	pod3.Name = "pod3"
	currentPodList := []corev1.Pod{
		pod1,
		pod2,
	}
	newPodList := []corev1.Pod{
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
