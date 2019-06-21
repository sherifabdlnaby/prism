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
	name           string
	bucket         string
	config         config.Pipeline
	status         status
	db             *bolt.DB
	Root           *node.Next
	NodeMap        map[string]*node.Node
	wg             sync.WaitGroup
	Logger         zap.SugaredLogger
	receiveTxnChan <-chan transaction.Transaction
	registry       registry.Registry
	hash           string
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

// Stop stops the pipeline, that means that any transaction received on this pipeline after stopping will return
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
func NewPipeline(name string, db *bolt.DB, Config config.Pipeline, registry registry.Registry, logger zap.SugaredLogger, hash string) (*Pipeline, error) {

	// Node Beginning Dummy Node
	root := node.NewNext(node.NewDummy("dummy", *resource.NewResource(Config.Concurrency), logger))

	// Create pipeline
	p := &Pipeline{
		name:           name,
		bucket:         name,
		config:         Config,
		status:         new,
		db:             db,
		Root:           root,
		NodeMap:        make(map[string]*node.Node),
		wg:             sync.WaitGroup{},
		Logger:         logger,
		receiveTxnChan: make(chan transaction.Transaction),
		registry:       registry,
		hash:           hash,
	}

	p.NodeMap[root.Name] = root.Node

	// Lookup Nexts of this Node
	nexts, err := getNexts(Config.Pipeline, p, false)
	if err != nil {
		return &Pipeline{}, nil
	}

	// set begin Node to nexts (Pipeline beginning)
	root.SetNexts(nexts)

	// create bucket for persistence
	err = p.createPersistenceBucket()
	if err != nil {
		return &Pipeline{}, nil
	}

	return p, nil
}

func getNexts(next map[string]*config.Node, p *Pipeline, forceSync bool) ([]node.Next, error) {
	nexts := make([]node.Next, 0)
	for Name, Node := range next {
		Node, err := buildTree(Name, *Node, p, forceSync)
		if err != nil {
			return nil, err
		}

		// create a next wrapper
		next := node.NewNext(Node)

		// append to nexts
		nexts = append(nexts, *next)
	}
	return nexts, nil
}

func buildTree(name string, n config.Node, p *Pipeline, forceSync bool) (*node.Node, error) {

	// create node of the configure components
	currNode, err := chooseComponent(name, p, len(n.Next))
	if err != nil {
		return nil, err
	}

	//
	async := n.Async
	if forceSync {
		async = false
	}

	// set node async
	currNode.SetAsync(async)

	// all NEXT nodes to be sync if current is async.
	if async {
		forceSync = true
	}

	// add nexts
	nexts, err := getNexts(n.Next, p, forceSync)
	if err != nil {
		return nil, err
	}

	// set nexts
	currNode.SetNexts(nexts)

	return currNode, nil
}

func getUniqueNodeName(name string, NodesList map[string]*node.Node) string {
	for i := 0; ; {
		_, ok := NodesList[name]
		if !ok {
			return name
		}
		name += "_" + strconv.Itoa(i)
	}
}

func chooseComponent(name string, p *Pipeline, nextsCount int) (*node.Node, error) {
	var Node *node.Node

	// check if ProcessReadWrite(and which types)
	processorBase, ok := p.registry.GetProcessor(name)
	if ok {
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", name)
		}
		switch base := processorBase.Base.(type) {
		case processor.ReadOnly:
			Node = node.NewReadOnly(name, base, processorBase.Resource, p.Logger)
		case processor.ReadWrite:
			Node = node.NewReadWrite(name, base, processorBase.Resource, p.Logger)
		case processor.ReadWriteStream:
			Node = node.NewReadWriteStream(name, base, processorBase.Resource, p.Logger)
		default:
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
		// Not ProcessReadWrite, check if output.
	} else {
		output, ok := p.registry.GetOutput(name)
		if ok {
			if nextsCount > 0 {
				return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", name)
			}
			Node = node.NewOutput(name, output, p.Logger)
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	// Set Attrs
	// Give new node a unique name
	Node.Name = getUniqueNodeName(Node.Name, p.NodeMap)

	// create a logger
	Node.Logger = *p.Logger.Named(Node.Name)

	// create a logger
	Node.Db = p.db

	// set bucket Name
	Node.Bucket = p.bucket

	// save in global map
	p.NodeMap[Node.Name] = Node

	return Node, nil
}

func (p *Pipeline) createPersistenceBucket() error {
	p.bucket = p.name + p.hash
	return p.db.Update(func(tx *bolt.Tx) error {
		var err error

		_, err = tx.CreateBucket([]byte(p.bucket))

		if err != nil && err != bolt.ErrBucketExists {
			return fmt.Errorf("create bucket: %s", err)
		}

		if err == bolt.ErrBucketExists {
			return nil
		}

		p.Logger.Infof("new pipeline, created a persistence bucket for new pipeline %s (%s)", p.name, p.hash)

		return nil
	})
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

	p.Logger.Infof("re-applying %d async requests found", len(TxnList))
	for _, asyncTxn := range TxnList {
		// Delete entry and tmp file
		err = p.db.Update(func(tx *bolt.Tx) error {
			// get payload and data
			tmpFile, err := os.Open(asyncTxn.Filepath)
			if err != nil {
				return err
			}

			// Lookup Node
			Node := p.NodeMap[asyncTxn.Node]

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
