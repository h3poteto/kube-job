package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:           "kube-job",
	Short:         "Run one off job on kubernetes",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().StringP("config", "", "", "Kubernetes config file path (If you don't set it, use environment variables `KUBECONFIG`)")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose mode")
	_ = viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))

	RootCmd.AddCommand(
		runJobCmd(),
		versionCmd(),
	)
}

func generalConfig() (string, bool) {
	config := viper.GetString("config")
	if len(config) == 0 {
		config = os.Getenv("KUBECONFIG")
		if len(config) == 0 {
			config = "$HOME/.kube/config"
		}
	}
	return config, viper.GetBool("verbose")
}
