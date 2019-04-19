package mirror

import (
	"bytes"
	"io"
	"sync/atomic"
)

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

//////
//
//type Reader struct {
//	internal io.Reader
//	buffer   bytes.Buffer
//	error    error
//	eofTotal int64
//}
//
//func (r *Reader) Read() (read int, error error) {
//	buffer := make([]byte, bytes.MinRead)
//	read, error = r.internal.Read(buffer)
//	r.buffer.Write(buffer)
//	return read, error
//}
//
//type readerCloner struct {
//	*Reader
//	i int
//}
//
//func (r *Reader) NewReader() io.Reader {
//	return &readerCloner{
//		Reader: r,
//		i:      0,
//	}
//}
//
//func (c *readerCloner) Read(p []byte) (read int, error error) {
//	upperlimit := c.i + len(p)
//
//	if upperlimit > c.buffer.Len() {
//		// try to read more
//
//		upperlimit = c.buffer.Len()
//
//		// check if EOF
//		if int64(upperlimit) >= c.eofTotal {
//			error = c.error
//		}
//
//	}
//
//	copy(p, c.buffer.Bytes()[c.i:upperlimit])
//
//	read = upperlimit - c.i
//
//	c.i = upperlimit
//
//	return read, error
//}
