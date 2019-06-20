package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/boltdb/bolt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/registry"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	receiveTxnChan <-chan transaction.Transaction
	Root           node.Next
	NodesList      map[string]*node.Node
	Logger         zap.SugaredLogger
	wg             sync.WaitGroup
	status         status
	name           string
	db             *bolt.DB
}

type status int32

const (
	_              = iota // ignore first value by assigning to blank identifier
	new     status = 1 + iota
	started status = 1 + iota
	closed  status = 1 + iota
)

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	// start pipeline nodes
	err := p.Root.Start()
	if err != nil {
		return err
	}

	// set status = started (no need atomic here, just for sake of consistency)
	atomic.SwapInt32((*int32)(&p.status), int32(started))

	go func() {
		for value := range p.receiveTxnChan {
			go p.job(value)
		}
	}()

	return nil
}

//ApplyPersistedAsyncRequests checks pipeline's persisted unfinished transactions and re-apply them
func (p *Pipeline) ApplyPersistedAsyncRequests() error {

	TxnList := make([]transaction.Async, 0)

	err := p.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(p.name))
		err := b.ForEach(func(k, v []byte) error {
			asyncTxn := &transaction.Async{}
			err := json.Unmarshal(v, asyncTxn)
			if err != nil {
				return err
			}
			TxnList = append(TxnList, *asyncTxn)
			return nil
		})

		return err
	})

	if err != nil {
		p.Logger.Infow("error occurred while reading from from in-disk DB", "error", err.Error())
		return err
	}

	p.Logger.Infof("reapplying %d async requests found", len(TxnList))
	for _, asyncTxn := range TxnList {
		// Delete entry and tmp file
		err = p.db.Update(func(tx *bolt.Tx) error {
			// get payload and data
			tmpFile, err := os.Open(asyncTxn.Filepath)
			if err != nil {
				return err
			}

			// Get Node
			Node := p.NodesList[asyncTxn.Node]

			responseChan := make(chan response.Response)

			Node.ProcessTransaction(transaction.Transaction{
				Payload:      io.Reader(tmpFile),
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

			err = tmpFile.Close()
			if err != nil {
				return err
			}

			b := tx.Bucket([]byte(p.name))
			err = b.Delete([]byte(asyncTxn.ID))
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			p.Logger.Errorw("an error occurred while applying persisted async requests", "error", err.Error())
		}
		// Delete from filesystem
		err = os.Remove(asyncTxn.Filepath)
		if err != nil {
			return err
		}

	}

	return nil
}

//Stop stops the pipeline, that means that any transaction received on this pipeline after stopping will return
// error response unless re-started again.
func (p *Pipeline) Stop() error {
	atomic.SwapInt32((*int32)(&p.status), int32(closed))

	// Wait all running jobs to return
	p.wg.Wait()

	//Stop
	err := p.Root.Stop()
	if err != nil {
		return err
	}

	return nil
}

//SetTransactionChan Set the transaction chan pipeline will use to receive input
func (p *Pipeline) SetTransactionChan(tc <-chan transaction.Transaction) {
	p.receiveTxnChan = tc
}

func (p *Pipeline) job(txn transaction.Transaction) {
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

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(name string, db *bolt.DB, pc config.Pipeline, registry registry.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	pipelineResource := resource.NewResource(pc.Concurrency)

	// dummy Node is the start of every pipeline, and its nexts(s) are the pipeline starting nodes.
	beginNode := node.NewDummy(*pipelineResource)
	beginNode.Name = "start"

	// NodeMap will contain all nodes of the pipeline. (will be useful later.
	NodeMap := make(map[string]*node.Node)
	NodeMap[beginNode.Name] = beginNode

	nexts := make([]node.Next, 0)
	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, *value, registry, NodeMap, logger, false)
		if err != nil {
			return nil, err
		}

		Node.SetAsync(value.Async)

		// create a next wrapper
		next := *node.NewNext(Node)

		// gives the next's Node its InputTransactionChan, now owner of the 'next' owns closing the chan.
		Node.SetTransactionChan(next.TransactionChan)

		// append to nexts
		nexts = append(nexts, next)
	}

	beginNode.SetNexts(nexts)

	// give dummy Node its receive chan
	Next := node.NewNext(beginNode)

	// give Node its receive chan
	beginNode.SetTransactionChan(Next.TransactionChan)

	pip := Pipeline{
		receiveTxnChan: make(chan transaction.Transaction),
		Root:           *Next,
		NodesList:      NodeMap,
		Logger:         logger,
		status:         new,
		name:           name,
		db:             db,
	}

	err := pip.db.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	//Add bucket ref too all nodes
	for _, Node := range NodeMap {
		Node.Bucket = name
		Node.Db = pip.db
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registry.Registry, NodesList map[string]*node.Node, logger zap.SugaredLogger, forceSync bool) (*node.Node, error) {

	// create node of the configure components
	currNode, err := chooseComponent(name, registry, len(n.Next))
	if err != nil {
		return nil, err
	}

	currNode.Name = name
	currNode.Logger = *logger.Named(name)

	// Give new node a unique name
	for i := 0; ; {
		_, ok := NodesList[currNode.Name]
		if ok {
			currNode.Name += "_" + strconv.Itoa(i)
			continue
		}
		NodesList[currNode.Name] = currNode
		break
	}

	// add nexts
	nexts := make([]node.Next, 0)
	if n.Next != nil {
		for key, value := range n.Next {

			Node, err := buildTree(key, *value, registry, NodesList, logger, n.Async)
			if err != nil {
				return nil, err
			}

			// create a next wrapper
			next := *node.NewNext(Node)

			// gives the next's node its InputTransactionChan, now owner of the 'next' owns closing the chan.
			Node.SetTransactionChan(next.TransactionChan)

			// append to nexts
			nexts = append(nexts, next)
		}
	}

	// set nexts
	currNode.SetNexts(nexts)

	// set node async
	currNode.SetAsync(n.Async && !forceSync)

	return currNode, nil
}

func chooseComponent(name string, registry registry.Registry, nextsCount int) (*node.Node, error) {
	var Node *node.Node

	// check if ProcessReadWrite(and which types)
	processorBase, ok := registry.GetProcessor(name)
	if ok {
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", name)
		}
		switch p := processorBase.Base.(type) {
		case processor.ReadOnly:
			Node = node.NewReadOnly(p, processorBase.Resource)
		case processor.ReadWrite:
			Node = node.NewReadWrite(p, processorBase.Resource)
		case processor.ReadWriteStream:
			Node = node.NewReadWriteStream(p, processorBase.Resource)
		default:
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
		// Not ProcessReadWrite, check if output.
	} else {
		output, ok := registry.GetOutput(name)
		if ok {
			if nextsCount > 0 {
				return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", name)
			}
			Node = node.NewOutput(output)
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	return Node, nil
}
