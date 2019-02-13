package job

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/ghodss/yaml"
	"k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type Job struct {
	client     *kubernetes.Clientset
	CurrentJob *v1.Job
}

func NewJob(configFile string, currentFile string, overrideJson string) (*Job, error) {
	if len(configFile) == 0 {
		return nil, errors.New("Config file is required")
	}
	if len(currentFile) == 0 {
		return nil, errors.New("Template file is required")
	}
	if len(overrideJson) == 0 {
		return nil, errors.New("Override json is required")
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

	// Patch override json does not offten contain kind and Version.
	// So I forgive to override JobSpec.
	var overrideJob v1.JobSpec
	if err := json.Unmarshal([]byte(overrideJson), &overrideJob); err != nil {
		return nil, err
	}

	return &Job{
		client,
		&currentJob,
		&overrideJob,
	}, nil
}

// func (j *Job) Run(namespace string, name string) {

// 	j.client.BatchV1().Jobs(namespace).Patch(name, types.JSONPatchType)
// }
