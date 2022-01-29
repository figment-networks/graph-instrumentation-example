package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/flags"
	"github.com/streamingfast/dlauncher/launcher"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sf-chain",
		Short: "Dummy chain on StreamingFast",
	}
)

func main() {
	cobra.OnInitialize(func() {
		//
		// This will allow setting flags via environment variables
		// Example: SF_GLOBAL_VERBOSE=3 will set `global-verbose` flag
		//
		flags.AutoBind(rootCmd, "SF")
	})

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		launcher.SetupLogger(&launcher.LoggingOptions{
			WorkingDir:    viper.GetString("global-data-dir"),
			Verbosity:     viper.GetInt("global-verbose"),
			LogFormat:     viper.GetString("global-log-format"),
			LogToFile:     viper.GetBool("global-log-to-file"),
			LogListenAddr: viper.GetString("global-log-level-switcher-listen-addr"),
		})
		return nil
	}

	rootCmd.AddCommand(
		initCommand,
		startCommand,
		setupCommand,
	)

	derr.Check("registering application flags", launcher.RegisterFlags(startCommand))
	derr.Check("executing root command", rootCmd.Execute())
}
