package libraryLogging

import (
	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
)

var logger *zap.Logger

var Info func(string, ...zap.Field)
var Warn func(string, ...zap.Field)
var Error func(string, ...zap.Field)
var Panic func(string, ...zap.Field)
var Fatal func(string, ...zap.Field)
var Debug func(string, ...zap.Field)

var Sync func() error

func init() {
	InitLogger()
}

func InitLogger(production ...bool) {
	if len(production) > 0 && production[0] {
		logger = zap.Must(zap.NewProduction())
	} else {
		logger = zap.Must(zap.NewDevelopment())
	}
	Info = logger.Info
	Warn = logger.Warn
	Error = logger.Error
	Panic = logger.Panic
	Fatal = logger.Fatal
	Debug = logger.Debug
	Sync = logger.Sync
}

func Dumper(args ...any) {
	spew.Dump(args...)
}
