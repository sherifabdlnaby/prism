package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/mirror"
	"github.com/sherifabdlnaby/prism/pkg/payload"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Node A Node in the pipeline
type Node struct {
	Name           string
	async          bool
	nexts          []Next
	Bucket         string
	nodeType       component
	Db             *bolt.DB
	wg             sync.WaitGroup
	resource       resource.Resource
	Logger         zap.SugaredLogger
	receiveTxnChan chan transaction.Transaction //TODO make a receive only for more sanity
}

//Next Wraps the next Node plus the channel used to communicate with this Node to send input transactions.
type Next struct {
	*Node
	TransactionChan chan transaction.Transaction
}

//TODO rename
type component interface {
	job(t transaction.Transaction)
	jobStream(t transaction.Transaction)
}

//NewNext Create a new Next Node with the supplied Node.
func NewNext(node *Node) *Next {
	transactionChan := make(chan transaction.Transaction)

	// gives the next's Node its InputTransactionChan, now owner of the 'next' owns closing the chan.
	node.SetTransactionChan(transactionChan)

	return &Next{
		Node:            node,
		TransactionChan: transactionChan,
	}
}

func newBase(nodeType component, resource resource.Resource) *Node {
	return &Node{
		async:    false,
		wg:       sync.WaitGroup{},
		nexts:    nil,
		resource: resource,
		nodeType: nodeType,
	}
}

// Start starts this Node and all its next nodes to start receiving transactions
// By starting all next nodes, start async request handler, and start receiving transactions
func (n *Node) Start() error {
	// Start next nodes
	for _, value := range n.nexts {
		err := value.Start()
		if err != nil {
			return err
		}
	}

	go func() {
		for t := range n.receiveTxnChan {
			n.handleTransaction(t)
		}
	}()

	return nil
}

//Stop Stop this Node and stop all its next nodes.
func (n *Node) Stop() error {
	//wait async jobs to finish
	n.wg.Wait()

	for _, value := range n.nexts {
		// close this next-node chan
		close(value.TransactionChan)

		// tell this next-node to stop which in turn will close all its next(s) too.
		err := value.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

//SetTransactionChan Set the transaction chan Node will use to receive input
func (n *Node) SetTransactionChan(tc chan transaction.Transaction) {
	n.receiveTxnChan = tc
}

//ReceiveTxnChan Getter for receiveChan
func (n *Node) ReceiveTxnChan() <-chan transaction.Transaction {
	return n.receiveTxnChan
}

//SetNexts Set this Node's next nodes.
func (n *Node) SetNexts(nexts []Next) {
	n.nexts = nexts
}

//SetAsync Set if this Node is sync/async
func (n *Node) SetAsync(async bool) {
	n.async = async
}

// ProcessTransaction transaction according to its type stream/bytes
func (n *Node) ProcessTransaction(t transaction.Transaction) {
	// Start Job according to transaction payload Type
	switch t.Payload.(type) {
	case payload.Bytes:
		go n.nodeType.job(t)
	case payload.Stream:
		go n.nodeType.jobStream(t)
	default:
		// This theoretically shouldn't happen
		t.ResponseChan <- response.Error(fmt.Errorf("invalid transaction payload type, must be payload.Bytes or payload.Stream"))
	}
}

func (n *Node) handleTransaction(t transaction.Transaction) {
	// if Node is set async, convert to async transaction
	if n.async {
		err := n.startAsyncTransaction(&t)
		if err != nil {
			t.ResponseChan <- response.Error(err)
			return
		}
	}

	n.ProcessTransaction(t)
}

func (n *Node) startAsyncTransaction(t *transaction.Transaction) error {

	dirPath := config.PRISM_TMP_DIR.Lookup()

	tmpFile, err := ioutil.TempFile(dirPath, "*.bat")
	if err != nil {
		return err
	}

	filepath := tmpFile.Name()

	//----------------------------------------------------------

	var newPayload payload.Payload

	// Lookup all Transaction Data
	switch Payload := t.Payload.(type) {
	case payload.Bytes:
		nBytes, err := tmpFile.Write(Payload)

		if err != nil {
			err = fmt.Errorf("failed to save async transaction to tmp tmpFile, error: %s", err.Error())
			return err
		}

		if nBytes != len(Payload) {
			err = fmt.Errorf("failed to save async transaction to tmp tmpFile, couldn't write all bytes")
			return err
		}

		// set new Payload to bytes (we keep in memory if we're in memory)
		newPayload = payload.Bytes(Payload)

		err = tmpFile.Close()
		if err != nil {
			return err
		}

		tmpFile = nil
	case payload.Stream:
		_, err = io.Copy(tmpFile, Payload)

		if err != nil {
			_ = tmpFile.Close()
			return err
		}

		err = tmpFile.Close()
		if err != nil {
			return err
		}

		tmpFile, err = os.Open(tmpFile.Name())

		newPayload = payload.Stream(tmpFile)

		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf("invalid transaction Payload type, must be Payload.Bytes or Payload.Stream")
		return err
	}

	// ----------------------------------------------------------

	asyncTxn := &transaction.Async{
		ID:       uuid.New().String(),
		Filepath: filepath,
		Pipeline: n.Bucket,
		Node:     n.Name,
		Data:     t.Data,
	}

	encodedBytes, err := json.Marshal(asyncTxn)
	if err != nil {
		return err
	}

	// Persist to Database
	err = n.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(n.Bucket))
		err := b.Put([]byte(asyncTxn.ID), encodedBytes)
		return err
	})

	if err != nil {
		return err
	}

	// --------------------------------------------------------------

	newResponseChan := make(chan response.Response)

	n.wg.Add(1)
	go n.receiveAsyncResponse(asyncTxn.ID, tmpFile, newResponseChan)

	// since it will be async, sync transaction context is irrelevant.
	// (we don't want sync nodes -that cancel transaction context when finishing to avoid ctx leak- to
	// cancel async nodes too )
	t.Context = context.Background()

	// New Payload
	t.Payload = newPayload

	// return ack response
	t.ResponseChan <- response.Ack()

	// now actual response is given to asyncResponds that should handle async responds
	t.ResponseChan = newResponseChan

	return nil
}

func (n *Node) receiveAsyncResponse(ID string, TmpFile *os.File, newResponseChan chan response.Response) {
	defer n.wg.Done()

	//TODO check Response
	response := <-newResponseChan
	if response.Error != nil {
		n.Logger.Errorw("error occurred when processing an async request", "error", response.Error.Error())
	}

	//close Tmp file if it was used to stream data
	if TmpFile != nil {
		_ = TmpFile.Close()
	}

	tmpFilePath := ""

	// DELETE TMP FILE AND DATABASE ENTRY
	err := n.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(n.Bucket))
		result := b.Get([]byte(ID))
		if result == nil {
			return fmt.Errorf("received response of a non persistent transaction")
		}

		asyncTxn := &transaction.Async{}
		err := json.Unmarshal(result, asyncTxn)
		if err != nil {
			return err
		}

		err = b.Delete([]byte(ID))
		if err != nil {
			return err
		}

		tmpFilePath = asyncTxn.Filepath

		return nil
	})
	if err != nil {
		return
	}

	// Delete from Storage
	err = os.Remove(tmpFilePath)
	if err != nil {
		return
	}
}

func (n *Node) sendNextsStream(ctx context.Context, writerCloner mirror.Cloner, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.TransactionChan <- transaction.Transaction{
			Payload:      writerCloner.Clone(),
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) sendNexts(ctx context.Context, output payload.Bytes, data payload.Data) chan response.Response {
	responseChan := make(chan response.Response, len(n.nexts))

	for _, next := range n.nexts {
		// Copy new map
		newData := make(payload.Data, len(data))
		for key := range data {
			newData[key] = data[key]
		}

		next.TransactionChan <- transaction.Transaction{
			Payload:      output,
			Data:         newData,
			Context:      ctx,
			ResponseChan: responseChan,
		}
	}
	return responseChan
}

func (n *Node) waitResponses(responseChan chan response.Response) response.Response {
	////////////////////////////////////////////
	// receive from next nodes
	count, total := 0, len(n.nexts)
	Response := response.Response{}

	for ; count < total; count++ {
		Response = <-responseChan
		if !Response.Ack {
			break
		}
	}

	return Response
}
