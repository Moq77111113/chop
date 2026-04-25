package main

import (
	"os"

	"github.com/spf13/cobra"
)

const exitFailure = 1

var rootCmd = &cobra.Command{
	Use:           "chop",
	Short:         "Local streaming/network impairment testbench for video pipelines",
	SilenceUsage:  true,
	SilenceErrors: false,
}

// Execute assembles the command tree and runs it. Exits with code 1 on
// error; cobra has already printed the message to stderr.
func Execute() {
	rootCmd.AddCommand(runCmd, lintCmd, blockCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitFailure)
	}
}
