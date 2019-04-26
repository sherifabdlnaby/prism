package node

import (
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

//DummyNode Used at the start of every pipeline.
type DummyNode struct {
}

//Job Just forwards the input.
func (dn *DummyNode) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, response.Response) {
	readerCloner := mirror.NewReader(t.Payload.Reader)
	return readerCloner, t.ImageBytes, response.ACK
}
