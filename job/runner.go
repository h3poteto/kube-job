package job

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func (j *Job) Run() error {
	running, err := j.RunJob()
	if err != nil {
		return err
	}
	log.Infof("Starting job: %s", running.Name)
	ctx, cancel := context.WithCancel(context.Background())
	if j.Timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), j.Timeout)
	}
	defer cancel()

	err = j.WaitJob(ctx, running)
	return err
}
