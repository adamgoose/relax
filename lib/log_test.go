package lib

import (
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	lc := zap.NewDevelopmentConfig()
	lc.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, _ := lc.Build()
	log = l.Named("test").Sugar()
}
