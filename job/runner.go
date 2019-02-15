/*
Package job provides simple functions to run a job on kubernetes.

Usage:
    import "github.com/h3poteto/kube-job/job"

Run a job overriding the commands

When you want to run a job on kubernetes, please use this package as follows.

At first, you have to prepare yaml for job, and provide a command to override the yaml.

For example:

    j, err := job.NewJob("$HOME/.kube/config", "job-template.yaml", "echo hoge", "target-container-name", 0 * time.Second)
    if err != nil {
        return err
    }

    // Run the job
    running, err := j.RunJob()
    if err != nil {
        return err
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    err = j.WaitJob(ctx, running)

Polling the logs

You can polling the logs with stream.

For example:

    // j is a Job struct
    watcher := NewWatcher(j.client, j.Container)

    // running is a batchv1.Job struct
    err := watcher.Watch(running, ctx)
    if err != nil {
        return err
    }

*/
package job

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

// Run a command on kubernetes cluster, and watch the log.
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
