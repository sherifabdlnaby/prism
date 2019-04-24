package node

import (
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
)

type DummyNode struct {
}

func (dn *DummyNode) Job(t transaction.Transaction) (mirror.ReaderCloner, transaction.ImageBytes, transaction.Response) {
	readerCloner := mirror.NewReader(t.Payload.Reader)
	return readerCloner, t.ImageBytes, transaction.ResponseACK
}
