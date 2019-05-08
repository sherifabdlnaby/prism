package mirror

import (
	"bytes"
	"io"
	"sync"
)

/////////////////////////////////////////////////////////////////////////////

//Reader Allow to writerCloner a reader and be able to readMore it more than once (it's done by introducing a mid-buffer)
type Reader struct {
	reader        io.Reader
	buf           []byte
	error         error
	originalTotal int64
	mx            sync.Mutex
	stepSize      int
	curr          int
}

//Clone Create a new Reader
func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader, buf: make([]byte, bytes.MinRead), stepSize: bytes.MinRead / 2}
}

func (r *Reader) readMore() {
	if r.error != nil {
		return
	}

	// grow if needed
	l := len(r.buf)
	if r.curr+r.stepSize > l {
		buf := make([]byte, 2*l+r.stepSize)
		copy(buf, r.buf)
		r.buf = buf
	}

	var n int
	n, r.error = r.reader.Read(r.buf[r.curr : r.curr+r.stepSize])
	r.curr += n

	if r.error != nil {
		r.originalTotal = int64(r.curr)
	}

	r.stepSize *= 2

	return
}

type readerCloner struct {
	*Reader
	i int
}

//Clone Create a new cloned reader.
func (r *Reader) Clone() io.Reader {
	return &readerCloner{
		Reader: r,
		i:      0,
	}
}

func (c *readerCloner) Read(p []byte) (read int, error error) {
	upperlimit := c.i + len(p)

	c.mx.Lock()
	if upperlimit > c.curr {
		// try to readMore more
		c.readMore()

		// is upperlimit still over len after readMore?
		if upperlimit > c.curr {
			upperlimit = c.curr
		}

		// check if error
		if int64(upperlimit) >= c.originalTotal {
			error = c.error
		}

	}
	c.mx.Unlock()

	copy(p, c.buf[c.i:upperlimit])

	read = upperlimit - c.i

	c.i = upperlimit

	return read, error
}
