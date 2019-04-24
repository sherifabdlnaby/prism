package resource

import (
	"github.com/sherifabdlnaby/semaphore"
	"go.uber.org/zap"
)

//Resource contains types required to control access to a resource
type Resource struct {
	*semaphore.Weighted
	Logger zap.SugaredLogger
}

func NewResource(concurrency int, logger zap.SugaredLogger) *Resource {
	return &Resource{
		Weighted: semaphore.NewWeighted(int64(concurrency)),
		Logger:   logger,
	}
}
