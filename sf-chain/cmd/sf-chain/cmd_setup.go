package main

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	setupCommand = &cobra.Command{
		Use:   "setup",
		Short: "Configures and initializes the project files",
		RunE:  runSetupCommand,
	}
)

func runSetupCommand(cmd *cobra.Command, args []string) error {
	return errors.New("this command is not implemented")
}
