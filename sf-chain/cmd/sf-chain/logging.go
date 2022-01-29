package main

import (
	"github.com/streamingfast/dlauncher/launcher"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

var (
	userLog = launcher.UserLog
	zlog    *zap.Logger
)

func init() {
	logging.Register("main", &zlog)
	logging.Set(logging.MustCreateLogger())
}
