package pipeline

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/app/resource"
	processor2 "github.com/sherifabdlnaby/prism/pkg/component/processor"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	receiveTxnChan <-chan transaction.Transaction
	Root           node.Next
	NodesList      []node.Node
	Logger         zap.SugaredLogger
	wg             sync.WaitGroup
	status         status
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
func NewPipeline(pc config.Pipeline, registry registery.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	pipelineResource := resource.NewResource(pc.Concurrency)

	// dummy Node is the start of every pipeline, and its nexts(s) are the pipeline starting nodes.
	beginNode := node.NewDummy(*pipelineResource)

	// NodesList will contain all nodes of the pipeline. (will be useful later.
	NodesList := make([]node.Node, 0)
	NodesList = append(NodesList, beginNode)

	nexts := make([]node.Next, 0)
	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, *value, registry, &NodesList, false)
		if err != nil {
			return nil, err
		}

		Node.SetAsync(value.Async)

		// create a next wrapper
		next := *node.NewNext(Node)

		// gives the next's node its InputTransactionChan, now owner of the 'next' owns closing the chan.
		Node.SetTransactionChan(next.TransactionChan)

		// append to nexts
		nexts = append(nexts, next)
	}

	beginNode.SetNexts(nexts)

	// give dummy node its receive chan
	Next := node.NewNext(beginNode)

	// give node its receive chan
	beginNode.SetTransactionChan(Next.TransactionChan)

	pip := Pipeline{
		receiveTxnChan: make(chan transaction.Transaction),
		Root:           *Next,
		NodesList:      NodesList,
		Logger:         logger,
		status:         new,
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registery.Registry, NodesList *[]node.Node, forceSync bool) (node.Node, error) {

	// create node of the configure components
	currNode, err := chooseComponent(name, registry, len(n.Next))
	if err != nil {
		return nil, err
	}

	*NodesList = append(*NodesList, currNode)

	// add nexts
	nexts := make([]node.Next, 0)

	if n.Next != nil {
		for key, value := range n.Next {

			Node, err := buildTree(key, *value, registry, NodesList, n.Async)
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

func chooseComponent(name string, registry registery.Registry, nextsCount int) (node.Node, error) {
	var Node node.Node

	// check if ProcessReadWrite(and which types)
	processor, ok := registry.GetProcessor(name)
	if ok {
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", name)
		}
		switch p := processor.Base.(type) {
		case processor2.ReadOnly:
			Node = node.NewReadOnly(p, processor.Resource)
		case processor2.ReadWrite:
			Node = node.NewReadWrite(p, processor.Resource)
		case processor2.ReadWriteStream:
			Node = node.NewReadWriteStream(p, processor.Resource)
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
