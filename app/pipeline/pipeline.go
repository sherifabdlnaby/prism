package pipeline

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

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
func NewPipeline(name string, pc config.Pipeline, registry registry.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	pipelineResource := resource.NewResource(pc.Concurrency)

	// dummy Node is the start of every pipeline, and its nexts(s) are the pipeline starting nodes.
	beginNode := node.NewDummy(*pipelineResource)
	beginNode.Name = "start"

	// NodesList will contain all nodes of the pipeline. (will be useful later.
	NodesList := make(map[string]*node.Node, 0)
	NodesList[beginNode.Name] = beginNode

	nexts := make([]node.Next, 0)
	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, *value, registry, NodesList, false)
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
		name:           name,
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registry.Registry, NodesList map[string]*node.Node, forceSync bool) (*node.Node, error) {

	// create node of the configure components
	currNode, err := chooseComponent(name, registry, len(n.Next))
	if err != nil {
		return nil, err
	}

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
			Node = node.NewOutput(output).Node
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	Node.Name = name

	return Node, nil
}
