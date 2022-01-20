package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/figment-networks/sf-project-template/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
)

type IngestorApp struct {
	*shutter.Shutter
	logsDir string
}

func (app *IngestorApp) Run() error {
	zlog.Info("starting ingestor", zap.String("logs-dir", app.logsDir))
	defer zlog.Info("stopped ingestor")

	src, err := os.Open("<SOURCE>")
	if err != nil {
		return err
	}
	defer src.Close()

	lines := make(chan string)

	reader, err := codec.NewLogReader(lines, "DMLOG")
	if err != nil {
		return err
	}

	go func() {
		for {
			data, err := reader.Read()
			if err != nil && err != io.EOF {
				zlog.Error("log reader error", zap.Error(err))
				reader.Close()
				return
			}

			// TODO: process the data
			fmt.Println("Data:", data)
		}
	}()

	// TODO: Rework this
	buf := make([]byte, 50*1024*1024)
	scanner := bufio.NewScanner(src)
	scanner.Buffer(buf, 50*1024*1024)

	for scanner.Scan() {
		lines <- scanner.Text()
	}

	return nil
}

func init() {
	flags := func(cmd *cobra.Command) error {
		cmd.Flags().String("ingestor-mode", "logs", "mode of operation")
		cmd.Flags().String("ingestor-logs-dir", "", "directory where instrumentation logs are stored")
		cmd.Flags().String("ingestor-logs-pattern", ".log", "pattern of the log files")
		cmd.Flags().Bool("ingestor-logs-watch", true, "exit when all matched files are processed")

		return nil
	}

	initFunc := func(runtime *launcher.Runtime) (err error) {
		dir := viper.GetString("ingestor-logs-dir")
		if dir == "" {
			return errors.New("ingestor logs dir must be set")
		}

		dir, err = expandDir(dir)
		if err != nil {
			return err
		}

		if !dirExists(dir) {
			return errors.New("ingestor logs dir must exist")
		}

		return nil
	}

	factoryFunc := func(runtime *launcher.Runtime) (launcher.App, error) {
		return &IngestorApp{
			Shutter: shutter.New(),
		}, nil
	}

	launcher.RegisterApp(&launcher.AppDef{
		ID:            "ingestor",
		Title:         "Ingestor",
		Description:   "Reads the log files produces by the instrumented node",
		MetricsID:     "ingestor",
		Logger:        launcher.NewLoggingDef("ingestor.*", nil),
		RegisterFlags: flags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
}
