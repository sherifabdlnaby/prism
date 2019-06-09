package dummy

import "io"

type customReader struct {
	reader io.Reader
	buf    []byte
	off    int
}

// Creates a new customReader using r
func newCustomReader(r io.Reader) *customReader {
	return &customReader{r, []byte{}, 0}
}

// Creates a new customReader using a byte slice as a buffer
// and no actual reader
func newCustomReaderBuffer(b []byte) *customReader {
	return &customReader{nil, b, 0}
}

// Resets the offset to the beginning of the
// buffer, so you can reread the reader
func (r *customReader) Reset() {
	r.off = 0
}

func (r *customReader) readBuffer(p []byte) (int, error) {
	i := 0
	for ; r.off < len(r.buf) && i < len(p); r.off, i = r.off+1, i+1 {
		p[i] = r.buf[r.off]
	}
	return i, nil
}

func (r *customReader) Read(p []byte) (int, error) {
	a, _ := r.readBuffer(p)

	b := 0
	if a < len(p) {
		b, _ = r.reader.Read(p[a:])
		r.off += len(p[a:])
		r.buf = append(r.buf, p[a:]...)
	}

	return a + b, nil
}
