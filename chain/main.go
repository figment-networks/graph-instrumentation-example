package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/figment-networks/graph-instrumentation-example/chain/core"
	"github.com/figment-networks/graph-instrumentation-example/chain/deepmind"
)

var cliOpts = struct {
	logLevel  string
	storeDir  string
	blockRate int
}{}

func init() {
	flag.StringVar(&cliOpts.storeDir, "store-dir", "./data", "Directory to store blocks data")
	flag.IntVar(&cliOpts.blockRate, "block-rate", 1, "Number of blocks to produce per second")
	flag.StringVar(&cliOpts.logLevel, "log-level", "info", "Log level")
	flag.Parse()

	level, err := logrus.ParseLevel(cliOpts.logLevel)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	// A global flag to enable instrumentation
	if os.Getenv("DM_ENABLED") == "1" {
		deepmind.Enable(os.Stdout)
	}
}

func main() {
	node := core.NewNode(cliOpts.storeDir, cliOpts.blockRate)

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

func waitForSignal() os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGINT)
	return <-sig
}
