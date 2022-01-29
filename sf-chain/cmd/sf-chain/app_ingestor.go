package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/blockstream"
	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/node-manager/mindreader"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	pbheadinfo "github.com/streamingfast/pbgo/sf/headinfo/v1"
	"github.com/streamingfast/shutter"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pbcodec "github.com/figment-networks/graph-instrumentation-example/chain/proto"
	"github.com/figment-networks/graph-instrumentation-example/sf-chain/codec"
)

const (
	modeLogs  = "logs"
	modeStdin = "stdin"
)

func init() {
	registerFlags := func(cmd *cobra.Command) error {
		flags := cmd.Flags()

		flags.String("ingestor-mode", modeStdin, "mode of operation")
		flags.String("ingestor-logs-dir", "", "directory where instrumentation logs are stored")
		flags.String("ingestor-logs-pattern", ".log", "pattern of the log files")
		flags.Bool("ingestor-logs-watch", true, "exit when all matched files are processed")
		flags.Int("ingestor-line-buffer-size", 10*1024*1024, "line reader buffer size")
		flags.String("mindreader-node-working-dir", "{sf-data-dir}/workdir", "Path where mindreader will stores its files")

		return nil
	}

	initFunc := func(runtime *launcher.Runtime) (err error) {
		switch viper.GetString("ingestor-mode") {
		case modeLogs:
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
		sfDataDir := runtime.AbsDataDir

		oneBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-oneblock-store-url"))
		mergedBlockStoreURL := mustReplaceDataDir(sfDataDir, viper.GetString("common-blocks-store-url"))
		workingDir := mustReplaceDataDir(sfDataDir, viper.GetString("mindreader-node-working-dir"))
		//gprcListenAdrr := viper.GetString("mindreader-node-grpc-listen-addr")
		mergeAndStoreDirectly := viper.GetBool("mindreader-node-merge-and-store-directly")
		mergeThresholdBlockAge := viper.GetDuration("mindreader-node-merge-threshold-block-age")
		batchStartBlockNum := viper.GetUint64("mindreader-node-start-block-num")
		batchStopBlockNum := viper.GetUint64("mindreader-node-stop-block-num")
		waitTimeForUploadOnShutdown := viper.GetDuration("mindreader-node-wait-upload-complete-on-shutdown")
		oneBlockFileSuffix := viper.GetString("mindreader-node-oneblock-suffix")
		blocksChanCapacity := viper.GetInt("mindreader-node-blocks-chan-capacity")
		appLogger := zap.NewNop()

		tracker := bstream.NewTracker(50) // TODO: make a flag
		tracker.AddResolver(bstream.OffsetStartBlockResolver(1))

		consoleReaderFactory := func(lines chan string) (mindreader.ConsolerReader, error) {
			return codec.NewLogReader(lines, "")
		}

		consoleReaderTransformer := func(obj interface{}) (*bstream.Block, error) {
			return codec.BlockFromProto(obj.(*pbcodec.Block))
		}

		blockStreamServer := blockstream.NewUnmanagedServer(blockstream.ServerOptionWithLogger(appLogger))
		healthCheck := func(ctx context.Context) (isReady bool, out interface{}, err error) {
			return blockStreamServer.Ready(), nil, nil
		}

		server := dgrpc.NewServer2(
			dgrpc.WithLogger(appLogger),
			dgrpc.WithHealthCheck(dgrpc.HealthCheckOverGRPC|dgrpc.HealthCheckOverHTTP, healthCheck),
		)
		server.RegisterService(func(gs *grpc.Server) {
			pbheadinfo.RegisterHeadInfoServer(gs, blockStreamServer)
			pbbstream.RegisterBlockStreamServer(gs, blockStreamServer)
		})

		mrp, err := mindreader.NewMindReaderPlugin(
			oneBlockStoreURL,
			mergedBlockStoreURL,
			mergeAndStoreDirectly,
			mergeThresholdBlockAge,
			workingDir,
			consoleReaderFactory,
			consoleReaderTransformer,
			tracker,
			batchStartBlockNum,
			batchStopBlockNum,
			blocksChanCapacity,
			headBlockUpdater,
			func(error) {},
			true,
			waitTimeForUploadOnShutdown,
			oneBlockFileSuffix,
			blockStreamServer,
			appLogger,
		)
		if err != nil {
			log.Fatal("error initialising mind reader", zap.Error(err))
			return nil, nil
		}

		return &IngestorApp{
			Shutter:        shutter.New(),
			mrp:            mrp,
			mode:           viper.GetString("ingestor-mode"),
			lineBufferSize: viper.GetInt("ingestor-line-buffer-size"),
		}, nil
	}

	launcher.RegisterApp(&launcher.AppDef{
		ID:            "ingestor",
		Title:         "Ingestor",
		Description:   "Reads the log files produces by the instrumented node",
		MetricsID:     "ingestor",
		Logger:        launcher.NewLoggingDef("ingestor.*", nil),
		RegisterFlags: registerFlags,
		InitFunc:      initFunc,
		FactoryFunc:   factoryFunc,
	})
}

func headBlockUpdater(uint64, string, time.Time) {
	// TODO: will need to be implemented somewhere
}

type IngestorApp struct {
	*shutter.Shutter

	mode           string
	logsDir        string
	lineBufferSize int

	mrp *mindreader.MindReaderPlugin
}

func (app *IngestorApp) Run() error {
	zlog.Info("starting ingestor", zap.String("mode", app.mode))
	defer zlog.Info("stopped ingestor")

	zlog.Info("starting ingestor mind reader plugin")
	app.mrp.Launch()

	go func() {
		err := app.startScanner(os.Stdin)
		zlog.Info("stanner finished", zap.Error(err))
		app.mrp.Shutdown(err)
	}()

	<-app.mrp.Terminated()
	return nil
}

func (app *IngestorApp) startScanner(src io.Reader) error {
	scanner := bufio.NewReaderSize(os.Stdin, app.lineBufferSize)

	for {
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				zlog.Error("got an error from readstring", zap.Error(err))
				return err
			}

			if len(line) == 0 {
				zlog.Info("finished reading source")
				return nil
			}
		}

		app.mrp.LogLine(strings.TrimSpace(line))
	}
}
