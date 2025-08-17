package main

import (
	"github.com/devantler-tech/ksail-go/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCmd(version, commit, date)
  rootCmd.Execute()
}
