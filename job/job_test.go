package job

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	v1 "k8s.io/api/batch/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type mockedKubernetes struct {
	kubernetes.Interface
	mockedBatch batchv1.BatchV1Interface
	mockedCore  corev1.CoreV1Interface
}

type mockedBatchV1 struct {
	batchv1.BatchV1Interface
	mockedJob batchv1.JobInterface
}

type mockedJob struct {
	batchv1.JobInterface
	job *v1.Job
}

type mockedCoreV1 struct {
	corev1.CoreV1Interface
	mockedPod corev1.PodInterface
}

type mockedPod struct {
	corev1.PodInterface
	jobName string
	pod     *v1core.Pod
	podList *v1core.PodList
}

func (m mockedJob) Create(context.Context, *v1.Job, metav1.CreateOptions) (*v1.Job, error) {
	return m.job, nil
}

func (m mockedJob) Get(context.Context, string, metav1.GetOptions) (*v1.Job, error) {
	return m.job, nil
}

func (m mockedJob) Delete(context.Context, string, metav1.DeleteOptions) error {
	return nil
}

func (m mockedPod) DeleteCollection(ctx context.Context, deleteOptions metav1.DeleteOptions, options metav1.ListOptions) error {
	if options.LabelSelector != "job-name="+m.jobName {
		return errors.New("label does not match")
	}
	return nil
}

func (m mockedPod) List(ctx context.Context, options metav1.ListOptions) (*v1core.PodList, error) {
	return m.podList, nil
}

func (m mockedBatchV1) Jobs(namespace string) batchv1.JobInterface {
	return m.mockedJob
}

func (m mockedCoreV1) Pods(namespace string) corev1.PodInterface {
	return m.mockedPod
}

func (m mockedKubernetes) BatchV1() batchv1.BatchV1Interface {
	return m.mockedBatch
}

func (m mockedKubernetes) CoreV1() corev1.CoreV1Interface {
	return m.mockedCore
}

func TestRunJob(t *testing.T) {
	currentJob, err := readJobFromFile("../example/job.yaml")
	if err != nil {
		t.Error(err)
	}

	jobMock := mockedJob{
		job: currentJob,
	}
	batchV1Mock := mockedBatchV1{
		mockedJob: jobMock,
	}

	job := &Job{
		CurrentJob: currentJob,
		Args:       []string{"hoge", "fuga"},
		Container:  "alpine",
		Timeout:    10 * time.Minute,
		client: mockedKubernetes{
			mockedBatch: batchV1Mock,
		},
	}

	j, err := job.RunJob()
	if err != nil {
		t.Error(err)
	}
	if j.Name != currentJob.Name {
		t.Errorf("job create failed: %v", j)
	}
}

func TestCheckJobConditions(t *testing.T) {
	complete := []v1.JobCondition{
		v1.JobCondition{
			Type: "Complete",
		},
		v1.JobCondition{
			Type: "Complete",
		},
	}
	err := checkJobConditions(complete)
	if err != nil {
		t.Error(err)
	}

	failed := []v1.JobCondition{
		v1.JobCondition{
			Type: "Complete",
		},
		v1.JobCondition{
			Type: "Failed",
		},
	}
	err = checkJobConditions(failed)
	if err == nil {
		t.Error("should be failed")
	}
}

func TestWaitJobCompleteWithWaitAll(t *testing.T) {
	currentJob, err := readJobFromFile("../example/job.yaml")
	if err != nil {
		t.Error(err)
	}
	currentJob.Status.Active = 0
	currentJob.Status.Conditions = []v1.JobCondition{
		v1.JobCondition{
			Type: "Complete",
		},
	}
	jobMock := mockedJob{
		job: currentJob,
	}
	batchV1Mock := mockedBatchV1{
		mockedJob: jobMock,
	}
	job := &Job{
		CurrentJob: currentJob,
		Args:       []string{"hoge", "fuga"},
		Container:  "alpine",
		Timeout:    10 * time.Minute,
		client: mockedKubernetes{
			mockedBatch: batchV1Mock,
		},
	}
	ctx := context.Background()
	err = job.WaitJobComplete(ctx, currentJob, true)
	if err != nil {
		t.Error(err)
	}
}

func TestWaitJobCompleteForContainer(t *testing.T) {
	currentJob, err := readJobFromFile("../example/job.yaml")
	if err != nil {
		t.Error(err)
	}
	currentJob.Status.Active = 2
	jobMock := mockedJob{
		job: currentJob,
	}
	batchV1Mock := mockedBatchV1{
		mockedJob: jobMock,
	}

	successPod := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodSucceeded,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	successPod.Name = "success"
	successContainer := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodRunning,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	successContainer.Name = "success"
	podList := v1core.PodList{
		Items: []v1core.Pod{
			successPod,
			successContainer,
		},
	}
	podMock := mockedPod{
		podList: &podList,
	}
	coreV1Mock := mockedCoreV1{
		mockedPod: podMock,
	}
	job := &Job{
		CurrentJob: currentJob,
		Args:       []string{"hoge", "fuga"},
		Container:  "alpine",
		Timeout:    10 * time.Minute,
		client: mockedKubernetes{
			mockedBatch: batchV1Mock,
			mockedCore:  coreV1Mock,
		},
	}
	ctx := context.Background()
	err = job.WaitJobComplete(ctx, currentJob, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCompleteTargetContainer(t *testing.T) {
	successPod := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodSucceeded,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	successPod.Name = "success"
	runningPod := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodRunning,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Running: &v1core.ContainerStateRunning{},
					},
				},
			},
		},
	}
	runningPod.Name = "running"
	successContainer := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodRunning,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	successContainer.Name = "success"
	failedContainer := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodRunning,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 1,
						},
					},
				},
			},
		},
	}
	failedContainer.Name = "failed"
	anotherPod := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodRunning,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "debian",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	anotherPod.Name = "another"
	failedPod := v1core.Pod{
		Status: v1core.PodStatus{
			Phase: v1core.PodFailed,
			ContainerStatuses: []v1core.ContainerStatus{
				v1core.ContainerStatus{
					Name: "alpine",
					State: v1core.ContainerState{
						Terminated: &v1core.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	failedPod.Name = "failed"

	pods := []v1core.Pod{
		successPod,
		successContainer,
	}
	completed := completeTargetContainer(pods, "alpine")
	if completed != true {
		t.Error(errors.New("succeed container should be completed"))
	}

	pods = []v1core.Pod{
		successContainer,
		anotherPod,
	}
	completed = completeTargetContainer(pods, "alpine")
	if completed != true {
		t.Error(errors.New("another pod should be completed"))
	}

	pods = []v1core.Pod{
		successContainer,
		failedPod,
	}
	completed = completeTargetContainer(pods, "alpine")
	if completed != true {
		t.Error(errors.New("failed pod should be completed"))
	}

	pods = []v1core.Pod{
		successContainer,
		failedContainer,
	}
	if completed != true {
		t.Error(errors.New("failed container should be completed"))
	}

	pods = []v1core.Pod{
		runningPod,
		successContainer,
	}
	completed = completeTargetContainer(pods, "alpine")
	if completed == true {
		t.Error(errors.New("running pod should not be completed"))
	}
}

func readJobFromFile(file string) (*v1.Job, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var currentJob v1.Job
	err = yaml.Unmarshal(bytes, &currentJob)
	if err != nil {
		return nil, err
	}
	currentJob.SetName(generateRandomName(currentJob.Name))
	return &currentJob, nil
}

func TestRemovePods(t *testing.T) {
	currentJob, err := readJobFromFile("../example/job.yaml")
	if err != nil {
		t.Error(t)
	}
	podMock := mockedPod{
		jobName: currentJob.Name,
	}
	coreV1Mock := mockedCoreV1{
		mockedPod: podMock,
	}
	job := &Job{
		CurrentJob: currentJob,
		Args:       []string{"hoge", "fuga"},
		Container:  "alpine",
		Timeout:    10 * time.Minute,
		client: mockedKubernetes{
			mockedCore: coreV1Mock,
		},
	}
	ctx := context.Background()
	err = job.removePods(ctx)
	if err != nil {
		t.Error(err)
	}
}
