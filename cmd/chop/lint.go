package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moq77111113/chop/internal/scenario"
)

var lintCmd = &cobra.Command{
	Use:   "lint <scenario.yml>",
	Short: "Validate a scenario file",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if _, err := scenario.Load(args[0]); err != nil {
			return err
		}
		fmt.Println("ok")
		return nil
	},
}
