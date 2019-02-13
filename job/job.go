package job

import (
	"context"
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

type Job struct {
	client     *kubernetes.Clientset
	CurrentJob *v1.Job
	Commands   []string
	Timeout    time.Duration
}

func NewJob(configFile string, currentFile string, command string, timeout time.Duration) (*Job, error) {
	if len(configFile) == 0 {
		return nil, errors.New("Config file is required")
	}
	if len(currentFile) == 0 {
		return nil, errors.New("Template file is required")
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

	p := shellwords.NewParser()
	commands, err := p.Parse(command)
	if err != nil {
		return nil, err
	}

	return &Job{
		client,
		&currentJob,
		commands,
		timeout,
	}, nil
}

func (j *Job) RunJob() (*v1.Job, error) {
	currentJob := j.CurrentJob.DeepCopy()
	currentJob.Spec.Template.Spec.Containers[0].Command = j.Commands

	resultJob, err := j.client.BatchV1().Jobs(j.CurrentJob.Namespace).Create(currentJob)
	if err != nil {
		return nil, err
	}
	return resultJob, nil
}

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
		log.Info("Job is success")
	case <-ctx.Done():
		return errors.New("process timeout")
	}

	return nil
}

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

func checkJobConditions(conditions []v1.JobCondition) error {
	for _, condition := range conditions {
		if condition.Type == v1.JobFailed {
			return fmt.Errorf("Job failed: %s", condition.Reason)
		}
	}
	return nil
}
