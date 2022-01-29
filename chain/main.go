package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"

	"github.com/figment-networks/graph-instrumentation-example/chain/core"
	"github.com/figment-networks/graph-instrumentation-example/chain/deepmind"
)

var cliOpts = struct {
	GenesisHeight uint64 `long:"genesis-height" description:"Blockhain genesis height" default:"1"`
	LogLevel      string `long:"log-level" description:"Logging level" default:"info"`
	StoreDir      string `long:"store-dir" description:"Directory for storing blocks data" default:"./data"`
	BlockRate     int    `long:"block-rate" description:"Block production rate (per second)" default:"1"`
}{}

func main() {
	if _, err := flags.ParseArgs(&cliOpts, os.Args); err != nil {
		return
	}

	level, err := logrus.ParseLevel(cliOpts.LogLevel)
	if err != nil {
		logrus.Fatal(err)
	}

	if cliOpts.BlockRate < 1 {
		logrus.Fatal("block rate option must be greater than 1")
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if os.Getenv("DM_ENABLED") == "1" {
		initDeepMind()
		defer deepmind.Shutdown()
	}

	node := core.NewNode(
		cliOpts.StoreDir,
		cliOpts.BlockRate,
		cliOpts.GenesisHeight,
	)

	if err := node.Initialize(); err != nil {
		logrus.WithError(err).Fatal("node failed to initialize")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := waitForSignal()
		logrus.WithField("signal", sig).Info("shutting down")
		cancel()
	}()

	if err := node.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("node terminated with error")
	} else {
		logrus.Info("node terminated")
	}
}

func initDeepMind() {
	// A global flag to enable instrumentation
	dmOutput := os.Getenv("DM_OUTPUT")

	switch dmOutput {
	case "", "stdout", "STDOUT":
		deepmind.Enable(os.Stdout)
	case "stderr", "STDERR":
		deepmind.Enable(os.Stderr)
	default:
		dmFile, err := os.OpenFile(dmOutput, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0666)
		if err != nil {
			logrus.WithError(err).Fatal("cant open DM output file")
		}
		deepmind.Enable(dmFile)
	}
}

func waitForSignal() os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGINT)
	return <-sig
}
