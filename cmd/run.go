package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type runJob struct {
	command string
	timeout int
}

func runJobCmd() *cobra.Command {
	r := &runJob{}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a job on Kubernetes",
		Run:   r.run,
	}

	flags := cmd.Flags()
	flags.StringVar(&r.command, "command", "", "Command which you want to run")
	flags.IntVarP(&r.timeout, "timeout", "t", 0, "Timeout seconds")

	return cmd
}

func (r *runJob) run(cmd *cobra.Command, args []string) {
	config, verbose := generalConfig()
	if !verbose {
		log.SetLevel(log.InfoLevel)
	}
	log.Infof("Using config file: %s", config)
}
