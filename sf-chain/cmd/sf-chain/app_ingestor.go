package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"

	"github.com/figment-networks/graph-instrumentation-example/sf-chain/codec"
)

type IngestorApp struct {
	*shutter.Shutter
	logsDir string
}

func (app *IngestorApp) Run() error {
	zlog.Info("starting ingestor", zap.String("logs-dir", app.logsDir))
	defer zlog.Info("stopped ingestor")

	linesChan := make(chan string)

	reader, err := codec.NewLogReader(linesChan, "DMLOG")
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

	scanner := bufio.NewReaderSize(os.Stdin, 50*1024*1024)

	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			if len(line) == 0 {
				return err
			}
		}

		if len(line) > 0 {
			linesChan <- line[0 : len(line)-1]
		}
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
		switch viper.GetString("ingestor-mode") {
		case "logs":
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
