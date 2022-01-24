package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dlauncher/launcher"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Starts all services at once",
		RunE:  runStartCommand,
		Args:  cobra.ArbitraryArgs,
	}
)

func runStartCommand(cmd *cobra.Command, args []string) error {
	var err error
	tracker := bstream.NewTracker(50)

	modules := &launcher.Runtime{
		AbsDataDir: "./sf-data",
		Tracker:    tracker,
	}

	launch := launcher.NewLauncher(modules)
	userLog.Debug("launcher created")

	apps := launcher.ParseAppsFromArgs(args, startAppByDefault)

	if err := launch.Launch(apps); err != nil {
		return err
	}

	signalHandler := derr.SetupSignalHandler(viper.GetDuration("common-system-shutdown-signal-delay"))
	select {
	case <-signalHandler:
		userLog.Printf("Received termination signal, quitting")
		go launch.Close()
	case appID := <-launch.Terminating():
		if launch.Err() == nil {
			userLog.Printf("Application %s triggered a clean shutdown, quitting", appID)
		} else {
			userLog.Printf("Application %s shutdown unexpectedly, quitting", appID)
			err = launch.Err()
		}
	}

	launch.WaitForTermination()
	return err
}

func startAppByDefault(app string) bool {
	return true
}
