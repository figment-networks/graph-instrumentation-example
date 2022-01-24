package main

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	initCommand = &cobra.Command{
		Use:   "init",
		Short: "Initialize local configuration",
		RunE:  runInitCommand,
	}
)

func runInitCommand(cmd *cobra.Command, args []string) error {
	return errors.New("this command is not implemented")
}
