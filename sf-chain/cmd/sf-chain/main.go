package main

import (
	"github.com/spf13/cobra"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/flags"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var (
	userLog = launcher.UserLog
	zlog    *zap.Logger

	rootCmd = &cobra.Command{
		Use:   "sf-project",
		Short: "Project on StreamingFast",
	}
)

func init() {
	logging.Register("main", &zlog)
	logging.Set(logging.MustCreateLogger())

	launcher.SetupLogger(&launcher.LoggingOptions{
		WorkingDir:    "./data",
		Verbosity:     3,
		LogFormat:     "text",
		LogToFile:     false,
		LogListenAddr: "localhost:4444",
	})
}

func main() {
	cobra.OnInitialize(func() {
		flags.AutoBind(rootCmd, "SF")
	})

	rootCmd.AddCommand(
		initCommand,
		startCommand,
		setupCommand,
	)

	derr.Check("registering application flags", launcher.RegisterFlags(startCommand))
	derr.Check("executing root command", rootCmd.Execute())
}
