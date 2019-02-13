package cmd

import (
	"github.com/h3poteto/kube-job/job"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runJob struct {
	command      string
	templateFile string
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
	flags.StringVarP(&r.command, "command", "c", "", "Command which you want to run")
	flags.IntVarP(&r.timeout, "timeout", "t", 0, "Timeout seconds")
	flags.BoolVar(&r.cleanup, "cleanup", false, "Celanup completed job after run the job")

	return cmd
}

func (r *runJob) run(cmd *cobra.Command, args []string) {
	config, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.InfoLevel)
	}
	log.Infof("Using config file: %s", config)
	j, err := job.NewJob(config, r.templateFile, r.command)
	if err != nil {
		log.Fatal(err)
	}
	if err := j.Run(); err != nil {
		log.Fatal(err)
	}
}
