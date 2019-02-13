package cmd

import (
	"github.com/h3poteto/kube-job/job"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runJob struct {
	patch        string
	templateFile string
	timeout      int
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
	flags.StringVarP(&r.patch, "patch", "p", "", "JSON which you want to override the template file")
	flags.IntVarP(&r.timeout, "timeout", "t", 0, "Timeout seconds")

	return cmd
}

func (r *runJob) run(cmd *cobra.Command, args []string) {
	config, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.InfoLevel)
	}
	j, err := job.NewJob(config, r.templateFile, r.patch)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Using config file: %s", config)
	log.Infof("currentJob: %+v", j.CurrentJob)
}
