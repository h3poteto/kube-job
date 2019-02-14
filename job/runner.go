package job

import (
	"context"
	"time"

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

	watcher := NewWatcher(j.client, j.Container)
	go func() {
		err := watcher.Watch(running, ctx)
		if err != nil {
			log.Error(err)
		}
	}()

	err = j.WaitJob(ctx, running)
	time.Sleep(10 * time.Second)
	cancel()
	return err
}
