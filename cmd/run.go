package cmd

import (
	"time"

	"github.com/h3poteto/kube-job/pkg/job"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runJob struct {
	args          string
	templateFile  string
	image         string
	container     string
	timeout       int
	cleanup       string
	ignoreSidecar bool
}

func runJobCmd() *cobra.Command {
	r := &runJob{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a job on Kubernetes",
		Run:   r.run,
	}

	flags := cmd.Flags()
	flags.StringVarP(&r.templateFile, "template-file", "f", "", "Job template file")
	flags.StringVar(&r.args, "args", "", "Command which you want to run")
	flags.StringVar(&r.image, "image", "", "Image which you want to run")
	flags.StringVar(&r.container, "container", "", "Container name which you want watch the log")
	flags.IntVarP(&r.timeout, "timeout", "t", 0, "Timeout seconds")
	flags.StringVar(&r.cleanup, "cleanup", "all", "Cleanup completed job after run the job. You can specify 'all', 'succeeded' or 'failed'.")
	flags.BoolVar(&r.ignoreSidecar, "ignore-sidecar", false, "Wait until all containers stop. If you set false, wait only specified container.")

	return cmd
}

func (r *runJob) run(cmd *cobra.Command, args []string) {
	config, verbose := generalConfig()
	log.SetLevel(log.DebugLevel)
	if !verbose {
		log.SetLevel(log.WarnLevel)
	}
	if r.cleanup != job.All.String() && r.cleanup != job.Succeeded.String() && r.cleanup != job.Failed.String() {
		err := errors.New("please set 'all', 'succeeded' or 'failed' as --cleanup")
		log.Fatal(err)
	}

	log.Infof("Using config file: %s", config)
	j, err := job.NewJob(config, r.templateFile, r.args, r.image, r.container, (time.Duration(r.timeout) * time.Second))
	if err != nil {
		log.Fatal(err)
	}

	if err := j.RunAndCleanup(r.cleanup, r.ignoreSidecar); err != nil {
		log.Fatal(err)
	}

}
