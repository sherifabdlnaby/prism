package component

import (
	"context"

	"github.com/sherifabdlnaby/semaphore"
)

//Resource contains types required to control access to a resource
type Resource struct {
	sema *semaphore.Weighted
}

//NewResource resource Constructor
func NewResource(concurrency int) *Resource {
	return &Resource{
		sema: semaphore.NewWeighted(int64(concurrency)),
	}
}

// Acquire resource, will block until resource is acquired or context is canceled
func (r *Resource) Acquire(ctx context.Context) error {
	// As Semaphore ignore ctx if there is available n in the semaphore, and Resource acquire is used to 'short-circuit'
	// so we will explicitly check for ctx.Done() before acquiring semaphore
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return r.sema.Acquire(ctx, 1)
	}
}

// Release resource
func (r *Resource) Release() {
	r.sema.Release(1)
}
