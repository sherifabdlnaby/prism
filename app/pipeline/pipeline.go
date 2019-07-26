package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/app/registry"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"io"
	"sync"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	name           string
	hash           string
	registry       registry.Registry
	Root           *node.Next
	NodeMap        map[string]*node.Node
	receiveTxnChan <-chan transaction.Transaction
	persistence    persistence.Persistence
	wg             sync.WaitGroup
	config         config.Pipeline
	Logger         zap.SugaredLogger
}

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	// start pipeline nodes
	err := p.Root.Start()
	if err != nil {
		return err
	}

	go func() {
		for value := range p.receiveTxnChan {
			go p.process(value)
		}
	}()

	return nil
}

// Stop stops the pipeline, that means that any transaction received on this pipeline after stopping will return
// error response unless re-started again.
func (p *Pipeline) Stop() error {
	// Wait all running jobs to return
	p.wg.Wait()

	//Stop
	err := p.Root.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) process(txn transaction.Transaction) {
	p.wg.Add(1)
	responseChan := make(chan response.Response)
	p.Root.TransactionChan <- transaction.Transaction{
		Payload:      txn.Payload,
		Data:         txn.Data,
		ResponseChan: responseChan,
		Context:      txn.Context,
	}
	txn.ResponseChan <- <-responseChan
	p.wg.Done()
}

//ApplyPersistedAsyncRequests checks pipeline's persisted unfinished transactions and re-apply them
func (p *Pipeline) ApplyPersistedAsyncRequests() error {

	TxnList, err := p.persistence.GetAllTxn()
	if err != nil {
		p.Logger.Infow("error occurred while reading in-disk transactions", "error", err.Error())
		return err
	}

	p.Logger.Infof("re-applying %d async requests found", len(TxnList))
	for _, asyncTxn := range TxnList {
		p.applyAsyncTxn(asyncTxn)
	}

	return nil
}

func (p *Pipeline) applyAsyncTxn(asyncTxn transaction.Async) {

	// Send Transaction to the Async Node
	responseChan := make(chan response.Response)
	p.NodeMap[asyncTxn.Node].ProcessTransaction(transaction.Transaction{
		Payload:      io.Reader(asyncTxn.TmpFile),
		Data:         asyncTxn.Data,
		Context:      context.Background(),
		ResponseChan: responseChan,
	})

	// Wait Response
	response := <-responseChan

	// log progress
	if !response.Ack {
		if response.Error != nil {
			p.Logger.Warnw("an async request that are re-done failed", "error", response.Error)
		} else if response.AckErr != nil {
			p.Logger.Warnw("an async request that are re-done was dropped", "reason", response.AckErr)
		}
	}

	// ------------------ Clean UP ------------------ //
	err := asyncTxn.Finalize()
	if err != nil {
		p.Logger.Errorw("an error occurred while applying finalizing async requests", "error", err.Error())
	}

	// Delete Entry from DB
	err = p.persistence.DeleteTxn(&asyncTxn)
	if err != nil {
		p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
	}
}
