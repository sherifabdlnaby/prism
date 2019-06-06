package app

import (
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

type pipelineWrapper struct {
	*pipeline.Pipeline
	TransactionChan       chan transaction.Transaction
	StreamTransactionChan chan transaction.Streamable
}

type logger struct {
	zap.SugaredLogger
	inputLogger      zap.SugaredLogger
	processingLogger zap.SugaredLogger
	outputLogger     zap.SugaredLogger
	pipelineLogger   zap.SugaredLogger
}

func newLoggers(c config.Config) *logger {
	l := logger{}
	l.SugaredLogger = c.Logger
	l.inputLogger = *l.SugaredLogger.Named("input")
	l.processingLogger = *l.SugaredLogger.Named("processor")
	l.outputLogger = *l.SugaredLogger.Named("output")
	l.pipelineLogger = *l.SugaredLogger.Named("pipeline")
	return &l
}
