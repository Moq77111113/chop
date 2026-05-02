package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/blocks/link"
	"github.com/moq77111113/chop/internal/blocks/process"
	"github.com/moq77111113/chop/internal/blocks/source"
)

const (
	blockTypeSource  = "source"
	blockTypeLink    = "link"
	blockTypeProcess = "process"
)

var builtinBlocks = map[string]block.Factory{
	blockTypeSource:  source.New,
	blockTypeLink:    link.New,
	blockTypeProcess: process.New,
}

var blockCmd = &cobra.Command{
	Use:                "block <type>",
	Short:              "Run a single block child (internal: invoked by the supervisor)",
	Hidden:             true,
	DisableFlagParsing: true,
	Args:               cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		typ := args[0]
		factory, ok := builtinBlocks[typ]
		if !ok {
			return fmt.Errorf("unknown block type: %s", typ)
		}
		block.RunBlock(typ, factory)
		return nil
	},
}
