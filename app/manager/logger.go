package manager

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"go.uber.org/zap"
)

type logger struct {
	baseLogger       zap.SugaredLogger
	inputLogger      zap.SugaredLogger
	processingLogger zap.SugaredLogger
	outputLogger     zap.SugaredLogger
	pipelineLogger   zap.SugaredLogger
}

func newLoggers(c config.Config) *logger {
	l := logger{}
	l.baseLogger = *c.Logger
	l.inputLogger = *l.baseLogger.Named("input")
	l.processingLogger = *l.baseLogger.Named("processor")
	l.outputLogger = *l.baseLogger.Named("output")
	l.pipelineLogger = *l.baseLogger.Named("pipeline")
	return &l
}
