package cmd

import (
	"time"

	"github.com/h3poteto/kube-job/job"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runJob struct {
	args         string
	templateFile string
	container    string
	timeout      int
	cleanup      bool
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
	flags.StringVar(&r.container, "container", "", "Container name which you want watch the log")
	flags.IntVarP(&r.timeout, "timeout", "t", 0, "Timeout seconds")
	flags.BoolVar(&r.cleanup, "cleanup", true, "Cleanup completed job after run the job if the job is succeeded")

	return cmd
}

func (r *runJob) run(cmd *cobra.Command, args []string) {
	config, verbose := generalConfig()
	log.SetLevel(log.DebugLevel)
	if !verbose {
		log.SetLevel(log.WarnLevel)
	}
	log.Infof("Using config file: %s", config)
	j, err := job.NewJob(config, r.templateFile, r.args, r.container, (time.Duration(r.timeout) * time.Second))
	if err != nil {
		log.Fatal(err)
	}
	if err := j.Run(); err != nil {
		log.Fatal(err)
	}
	if r.cleanup {
		if err := j.Cleanup(); err != nil {
			log.Fatal(err)
		}
	}
}
