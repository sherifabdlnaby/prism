package resource

import (
	"context"

	"github.com/sherifabdlnaby/semaphore"
)

//Resource contains types required to control access to a resource
type Resource struct {
	sema *semaphore.Weighted
}

//NewResource Resource Constructor
func NewResource(concurrency int) *Resource {
	return &Resource{
		sema: semaphore.NewWeighted(int64(concurrency)),
	}
}

// Acquire resource, will block until resource is acquired or context is canceled
func (r *Resource) Acquire(ctx context.Context) error {
	return r.sema.Acquire(ctx, 1)
}

// Release resource
func (r *Resource) Release() {
	r.sema.Release(1)
}
