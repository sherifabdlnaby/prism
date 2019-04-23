package resource

import (
	"github.com/sherifabdlnaby/semaphore"
)

//Resource contains types required to control access to a resource
type Resource struct {
	*semaphore.Weighted
}

func NewResource(concurrency int) *Resource {
	return &Resource{
		Weighted: semaphore.NewWeighted(int64(concurrency)),
	}
}

//
//
//func (*Resource) Acquire(context.Context, int64) error {
//	panic("implement me")
//}
//
//func (*Resource) TryAcquire(int64) bool {
//	panic("implement me")
//}
//
//func (*Resource) Release(int64) {
//	panic("implement me")
//}
//
//func (*Resource) Resize(int64) {
//	panic("implement me")
//}
