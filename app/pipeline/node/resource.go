package node

import (
	"github.com/sherifabdlnaby/semaphore"
	"go.uber.org/zap"
)

//Resource contains types required to control access to a resource
type Resource struct {
	*semaphore.Weighted
	Logger zap.SugaredLogger
}

//NewResource Resource Constructor
func NewResource(concurrency int, logger zap.SugaredLogger) *Resource {
	return &Resource{
		Weighted: semaphore.NewWeighted(int64(concurrency)),
		Logger:   logger,
	}
}
