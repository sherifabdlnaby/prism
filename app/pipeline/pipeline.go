package pipeline

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/registry"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	name           string
	hash           string
	status         status
	registry       registry.Registry
	Root           *node.Next
	NodeMap        map[string]*node.Node
	receiveTxnChan <-chan transaction.Transaction
	persistence    persistence.Persistence
	wg             sync.WaitGroup
	config         config.Pipeline
	Logger         zap.SugaredLogger
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
func NewPipeline(name string, Config config.Pipeline, registry registry.Registry, logger zap.SugaredLogger, hash string) (*Pipeline, error) {
	var err error

	// Node Beginning Dummy Node
	root := node.NewNext(node.NewDummy("dummy", resource.NewResource(Config.Concurrency), logger))

	// Create pipeline
	p := &Pipeline{
		name:           name,
		hash:           hash,
		status:         new,
		config:         Config,
		Root:           root,
		NodeMap:        make(map[string]*node.Node),
		receiveTxnChan: make(chan transaction.Transaction),
		wg:             sync.WaitGroup{},
		Logger:         logger,
		registry:       registry,
	}

	// create persistence
	p.persistence, err = persistence.NewPersistence(p.name, p.hash, p.Logger)
	if err != nil {
		return &Pipeline{}, err
	}

	// Set first node
	p.NodeMap[root.Name] = root.Node

	// Lookup Nexts of this Node
	nexts, err := getNexts(Config.Pipeline, p, false)
	if err != nil {
		return &Pipeline{}, err
	}

	// set begin Node to nexts (Pipeline beginning)
	root.SetNexts(nexts)

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

	//
	async := n.Async
	if forceSync {
		async = false
	}

	// create node of the configure components
	currNode, err := p.createNode(name, p.getUniqueNodeName(name), async, p.persistence, len(n.Next))
	if err != nil {
		return nil, err
	}

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

func (p Pipeline) getUniqueNodeName(name string) string {
	for i := 0; ; {
		_, ok := p.NodeMap[name]
		if !ok {
			return name
		}
		name += "_" + strconv.Itoa(i)
	}
}

func (p Pipeline) createNode(componentName, nodeName string, async bool, persistence persistence.Persistence, nextsCount int) (*node.Node, error) {
	var Node *node.Node

	// check if ProcessReadWrite(and which types)
	component := p.registry.GetComponent(componentName)
	if component == nil {
		return nil, fmt.Errorf("plugin [%s] doesn't exists", componentName)
	}

	switch component := component.(type) {
	case *wrapper.ProcessorReadWrite, *wrapper.ProcessorReadOnly, *wrapper.ProcessorReadWriteStream:
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", nodeName)
		}

		switch component := component.(type) {
		case *wrapper.ProcessorReadWrite:
			Node = node.NewReadWrite(nodeName, component, p.Logger)
		case *wrapper.ProcessorReadOnly:
			Node = node.NewReadOnly(nodeName, component, p.Logger)
		case *wrapper.ProcessorReadWriteStream:
			Node = node.NewReadWriteStream(nodeName, component, p.Logger)
		}

	case *wrapper.Output:
		if nextsCount > 0 {
			return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", nodeName)
		}
		Node = node.NewOutput(nodeName, component, p.Logger)
	default:
		return nil, fmt.Errorf("plugin [%s] doesn't exists", nodeName)
	}

	Node.SetAsync(async)

	Node.SetPersistence(p.persistence)

	// save in map
	p.NodeMap[nodeName] = Node

	return Node, nil
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
