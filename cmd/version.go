package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version  string
	revision string
	build    string
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version : %s\n", version)
			fmt.Printf("Revision: %s\n", revision)
			fmt.Printf("Build   : %s\n", build)
		},
	}

	return cmd
}
