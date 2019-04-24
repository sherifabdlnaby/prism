package mirror

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"
)

//TODO concurrency protection using RWMxS
//TODO performance Optimization for buffer allocations

type ReaderCloner interface {
	NewReader() io.Reader
}

//Writer A Writing buffer that can be used to create multiple readers that all can read *the same* data written.
type Writer struct {
	internal bytes.Buffer
	error    error
	eofTotal int64
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
func (r *Writer) Write(p []byte) (n int, err error) {
	n, err = r.internal.Write(p)
	if err != nil {
		r.error = err
	}
	return
}

//Close Signal that the writer should no longer accept any input. and return EOF to readers.
func (r *Writer) Close() error {

	atomic.SwapInt64(&r.eofTotal, int64(r.internal.Len()))
	r.error = io.EOF

	return nil
}

type writerCloner struct {
	*Writer
	i int
}

//NewReader A reader that can read all the data written.
func (r *Writer) NewReader() io.Reader {
	return &writerCloner{
		Writer: r,
		i:      0,
	}
}

func (c *writerCloner) Read(p []byte) (read int, error error) {
	upperlimit := c.i + len(p)

	if upperlimit > c.internal.Len() {
		upperlimit = c.internal.Len()

		// check if EOF
		if int64(upperlimit) >= c.eofTotal {
			error = c.error
		}

	}

	copy(p, c.internal.Bytes()[c.i:upperlimit])

	read = upperlimit - c.i

	c.i = upperlimit

	return read, error
}

////

//Reader Allow to clone a reader and be able to read it more than once (it's done by introducing a mid-buffer)
type Reader struct {
	reader   io.Reader
	internal bytes.Buffer
	buffer   []byte
	error    error
	eofTotal int64
	mx       sync.Mutex
}

//NewReader Create a new Reader
func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader}
}

func (r *Reader) read() {
	if r.buffer == nil {
		r.buffer = make([]byte, bytes.MinRead)
	}

	n, err := r.reader.Read(r.buffer)
	r.internal.Write(r.buffer[:n])

	if err == io.EOF {
		r.error = err
		r.eofTotal = int64(r.internal.Len())
	}

	return
}

type readerCloner struct {
	*Reader
	i int
}

//NewReader Create a new cloned reader.
func (r *Reader) NewReader() io.Reader {
	return &readerCloner{
		Reader: r,
		i:      0,
	}
}

func (c *readerCloner) Read(p []byte) (read int, error error) {
	upperlimit := c.i + len(p)

	c.mx.Lock()
	if upperlimit > c.internal.Len() {
		// try to read more
		c.read()

		// is upperlimit still over len after read?
		if upperlimit > c.internal.Len() {
			upperlimit = c.internal.Len()
		}

		// check if EOF
		if int64(upperlimit) >= c.eofTotal {
			error = c.error
		}

	}
	c.mx.Unlock()

	copy(p, c.internal.Bytes()[c.i:upperlimit])

	read = upperlimit - c.i

	c.i = upperlimit

	return read, error
}
