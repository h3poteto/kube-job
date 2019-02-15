package job

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"time"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Job has client of kubernetes, current job, command, timeout, and target container information.
type Job struct {
	client *kubernetes.Clientset

	// Batch v1 job struct.
	CurrentJob *v1.Job
	// Command which override the current job struct.
	Commands []string
	// Target container name.
	Container string
	// If you set 0, timeout is ignored.
	Timeout time.Duration
}

// NewJob returns a new Job struct, and initialize kubernetes client.
// It read the job definition yaml file, and unmarshal to batch/v1/Job.
func NewJob(configFile, currentFile, command, container string, timeout time.Duration) (*Job, error) {
	if len(configFile) == 0 {
		return nil, errors.New("Config file is required")
	}
	if len(currentFile) == 0 {
		return nil, errors.New("Template file is required")
	}
	if len(container) == 0 {
		return nil, errors.New("Container is required")
	}
	client, err := newClient(configFile)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile(currentFile)
	if err != nil {
		return nil, err
	}
	var currentJob v1.Job
	err = yaml.Unmarshal(bytes, &currentJob)
	if err != nil {
		return nil, err
	}
	currentJob.SetName(generateRandomName(currentJob.Name))

	p := shellwords.NewParser()
	commands, err := p.Parse(command)
	if err != nil {
		return nil, err
	}

	return &Job{
		client,
		&currentJob,
		commands,
		container,
		timeout,
	}, nil
}

func generateRandomName(name string) string {
	return fmt.Sprintf("%s-%s", name, secureRandomStr(16))
}

// secureRandomStr is generate random string.
func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := rand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}

// RunJob is run a kubernetes job, and returns the job information.
func (j *Job) RunJob() (*v1.Job, error) {
	currentJob := j.CurrentJob.DeepCopy()
	index, err := findContainerIndex(currentJob, j.Container)
	if err != nil {
		return nil, err
	}
	currentJob.Spec.Template.Spec.Containers[index].Command = j.Commands

	resultJob, err := j.client.BatchV1().Jobs(j.CurrentJob.Namespace).Create(currentJob)
	if err != nil {
		return nil, err
	}
	return resultJob, nil
}

// findContainerIndex finds target container from job definition.
func findContainerIndex(job *v1.Job, containerName string) (int, error) {
	for index, container := range job.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			return index, nil
		}
	}
	return 0, errors.New("Container does not exit in the template")
}

// WaitJob waits response of the job.
func (j *Job) WaitJob(ctx context.Context, job *v1.Job) error {
	log.Info("Waiting for running job...")

	errCh := make(chan error, 1)
	done := make(chan struct{}, 1)
	go func() {
		err := j.WaitJobComplete(job)
		if err != nil {
			errCh <- err
		}
		close(done)
	}()
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-done:
		log.Info("Job is succeeded")
	case <-ctx.Done():
		return errors.New("process timeout")
	}

	return nil
}

// WaitJobComplete waits the completion of the job.
// If the job is failed, this function returns error.
// If the job is succeeded, this function returns nil.
func (j *Job) WaitJobComplete(job *v1.Job) error {
retry:
	for {
		time.Sleep(3 * time.Second)
		running, err := j.client.BatchV1().Jobs(job.Namespace).Get(job.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if running.Status.Active == 0 {
			return checkJobConditions(running.Status.Conditions)
		}
		continue retry

	}
	return nil

}

// checkJobConditions checks conditions of all jobs.
// If any job is failed, returns error.
func checkJobConditions(conditions []v1.JobCondition) error {
	for _, condition := range conditions {
		if condition.Type == v1.JobFailed {
			return fmt.Errorf("Job is failed: %s", condition.Reason)
		}
	}
	return nil
}

// Cleanup removes the job from the kubernetes cluster.
func (j *Job) Cleanup() error {
	log.Info("Removing the job...")
	options := metav1.DeleteOptions{}
	return j.client.BatchV1().Jobs(j.CurrentJob.Namespace).Delete(j.CurrentJob.Name, &options)
}
