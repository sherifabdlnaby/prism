package bufferspool

import (
	"sync"
)

// TODO benchmark those sizes

// 5KB initial buffer size
var initialBufferSize = 5120

// 2MB max buffer size kept in Pool
var maxBufferSize = 2097152

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, initialBufferSize)
	},
}

func Get() []byte {
	return bufferPool.Get().([]byte)
}

func Put(buffer []byte) {
	if len(buffer) > maxBufferSize {
		return
	}
	bufferPool.Put(buffer)
}
