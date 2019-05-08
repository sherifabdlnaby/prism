package node

import (
	"sync"
)

// 5KB initial buffer size
var initialBufferSize = 5120

// 2MB max buffer size kept in Pool
// TODO benchmark
var maxBufferSize = 2097152

var buffersPool = bufferPool{
	pool: sync.Pool{
		New: func() interface{} {
			return make([]byte, initialBufferSize)
		},
	},
}

type bufferPool struct {
	pool sync.Pool
}

func (bp *bufferPool) Get() []byte {
	return bp.pool.Get().([]byte)
}

func (bp *bufferPool) Put(buffer []byte) {
	if len(buffer) > maxBufferSize {
		return
	}

	bp.pool.Put(buffer)
}
