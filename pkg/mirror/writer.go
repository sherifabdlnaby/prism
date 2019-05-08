package mirror

import (
	"io"
	"sync"
	"sync/atomic"
)

//Writer A Writing buffer that can be used to create multiple readers that all can readMore *the same* data written.
type Writer struct {
	buf      []byte
	curr     int
	error    error
	eofTotal int64
	mx       sync.Mutex
}

//NewWriter Returns a new Writer
func NewWriter(buffer []byte) *Writer {
	return &Writer{
		buf: buffer,
	}
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
func (r *Writer) Write(p []byte) (n int, err error) {
	r.mx.Lock()

	if r.curr+len(p) > len(r.buf) {
		buf := make([]byte, 2*len(r.buf)+len(p))
		copy(buf, r.buf)
		r.buf = buf
	}

	copy(r.buf[r.curr:], p)
	r.curr += len(p)

	r.mx.Unlock()
	return
}

//Close Signal that the writer should no longer accept any input. and return EOF to readers.
func (r *Writer) Close() error {
	atomic.SwapInt64(&r.eofTotal, int64(r.curr))
	r.error = io.EOF
	return nil
}

//Clone A reader that can readMore all the data written.
func (r *Writer) Clone() io.Reader {
	return &writerCloner{
		Writer: r,
	}
}

type writerCloner struct {
	*Writer
	i int
}

func (c *writerCloner) Read(p []byte) (read int, error error) {
	upperlimit := c.i + len(p)

	c.mx.Lock()
	if upperlimit > c.curr {
		upperlimit = c.curr

		// check if EOF
		if int64(upperlimit) >= c.eofTotal {
			error = c.error
		}

	}
	c.mx.Unlock()

	copy(p, c.buf[c.i:upperlimit])

	read = upperlimit - c.i

	c.i = upperlimit

	return read, error
}
