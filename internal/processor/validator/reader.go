package validator

import "io"

type reader struct {
	reader io.Reader
	buf    []byte
	off    int
}

func NewReader(r io.Reader) *reader {
	return &reader{r, []byte{}, 0}
}

func (r *reader) Reset() {
	r.off = 0
}

func (r *reader) readBuffer(p []byte) (int, error) {
	i := 0
	for ; r.off < len(r.buf) && i < len(p); r.off, i = r.off+1, i+1 {
		p[i] = r.buf[r.off]
	}
	return i, nil
}

func (r *reader) Read(p []byte) (int, error) {
	a, _ := r.readBuffer(p)

	b := 0
	if a < len(p) {
		b, _ = r.reader.Read(p[a:])
		r.off += len(p[a:])
		r.buf = append(r.buf, p[a:]...)
	}

	return a + b, nil
}
