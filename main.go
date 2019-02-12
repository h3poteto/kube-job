package main

import (
	"fmt"
	"os"

	"github.com/h3poteto/kube-job/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
