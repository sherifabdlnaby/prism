package mirror

import (
	"io"
)

// Cloner Allow to get create multiple readers from a reader, and each created reader will readMore the same data as
// base reader
type Cloner interface {
	Clone() io.Reader
}
