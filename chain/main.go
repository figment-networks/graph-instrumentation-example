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
}

func main() {
	if os.Getenv("DM_ENABLED") == "1" {
		initDeepMind()
		defer deepmind.Shutdown()
	}

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
