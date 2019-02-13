package job

import (
	"io/ioutil"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type Job struct {
	client     *kubernetes.Clientset
	CurrentJob *v1.Job
	Commands   []string
}

func NewJob(configFile string, currentFile string, command string) (*Job, error) {
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
	}, nil
}

func (j *Job) Run() error {
	currentJob := j.CurrentJob.DeepCopy()
	currentJob.Spec.Template.Spec.Containers[0].Command = j.Commands

	resultJob, err := j.client.BatchV1().Jobs(j.CurrentJob.Namespace).Create(currentJob)
	if err != nil {
		return err
	}
	log.Infof("Starting job: %v", *resultJob)
	return nil
}
