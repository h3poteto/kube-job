/*
Package job provides simple functions to run a job on kubernetes.

Usage:
    import "github.com/h3poteto/kube-job/pkg/job"

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

// CleanupType for enum.
type CleanupType int

const (
	// All is a clean up type. Remove the job and pods whether the job is succeeded or failed.
	All CleanupType = iota
	// Succeeded is a clean up type. Remove the job and pods when the job is succeeded.
	Succeeded
	// Failed is a cleanup type. Remove the job and pods when the job is failed.
	Failed
)

func (c CleanupType) String() string {
	return [...]string{"all", "succeeded", "failed"}[c]
}

// Run a command on kubernetes cluster, and watch the log.
func (j *Job) Run(ignoreSidecar bool) error {
	if ignoreSidecar {
		log.Info("Ignore sidecar containers")
	}
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

	err = j.WaitJob(ctx, running, ignoreSidecar)
	time.Sleep(10 * time.Second)
	return err
}

// RunAndCleanup executes a command and clean up the job and pods.
func (j *Job) RunAndCleanup(cleanupType string, ignoreSidecar bool) error {
	if err := j.Validate(); err != nil {
		return err
	}
	err := j.Run(ignoreSidecar)
	if !shouldCleanup(cleanupType, err) {
		log.Info("Job should no clean up")
		return err
	}
	if e := j.Cleanup(); e != nil {
		return e
	}
	return err
}

func shouldCleanup(cleanupType string, jobResult error) bool {
	return cleanupType == All.String() || (cleanupType == Succeeded.String() && jobResult == nil) || (cleanupType == Failed.String() && jobResult != nil)
}
