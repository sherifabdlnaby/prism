package pipeline

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/response"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	receiveChan <-chan transaction.Transaction
	Root        node.Next
	NodesList   []node.Node
	Logger      zap.SugaredLogger
	wg          sync.WaitGroup
	status      status
}

type status int32

const (
	_              = iota // ignore first value by assigning to blank identifier
	new     status = 1 + iota
	started status = 1 + iota
	closed  status = 1 + iota
)

//startMux starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	// start pipeline node back to front
	for i := range p.NodesList {
		err := p.NodesList[len(p.NodesList)-i-1].Start()
		if err != nil {
			return fmt.Errorf("failed to start pipeline: %s", err.Error())
		}
	}

	// set status = started (no need atomic here, just for sake of consistency)
	atomic.SwapInt32((*int32)(&p.status), int32(started))

	go func() {
		for value := range p.receiveChan {
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

	// Stop Nodes
	close(p.Root.TransactionChan)
	for _, Node := range p.NodesList {
		err := Node.Stop()
		if err != nil {
			return fmt.Errorf("failed to stop pipeline: %s", err.Error())
		}
	}

	return nil
}

func (n *Pipeline) SetTransactionChan(tc <-chan transaction.Transaction) {
	n.receiveChan = tc
}

func (p *Pipeline) job(txn transaction.Transaction) {
	p.wg.Add(1)
	responseChan := make(chan response.Response, 1)
	p.Root.TransactionChan <- transaction.Transaction{
		Payload:      txn.Payload,
		ImageData:    txn.ImageData,
		ResponseChan: responseChan,
		Context:      txn.Context,
	}
	txn.ResponseChan <- <-responseChan
	p.wg.Done()
}

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(pc config.Pipeline, registry registery.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	pipelineResource := resource.NewResource(pc.Concurrency, logger)

	// Dummy Node is the start of every pipeline, and its next(s) are the pipeline starting nodes.
	beginNode := node.Dummy{
		Resource: *pipelineResource,
	}

	// NodesList will contain all nodes of the pipeline. (will be useful later.
	NodesList := make([]node.Node, 0)
	NodesList = append(NodesList, &beginNode)

	next := make([]node.Next, 0)
	for key, value := range pc.Pipeline {
		Node, err := getNext(key, value, registry, &NodesList)
		if err != nil {
			pipelineResource.Logger.Error(err.Error())
			return &Pipeline{}, err
		}
		next = append(next, Node)
	}
	beginNode.Next = next

	// give dummy node its receive chan
	tc := make(chan transaction.Transaction)
	beginNode.SetTransactionChan(tc)

	pip := Pipeline{
		receiveChan: make(chan transaction.Transaction),
		Root: node.Next{
			Node:            &beginNode,
			TransactionChan: tc,
		},
		NodesList: NodesList,
		Logger:    logger,
		status:    new,
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registery.Registry, NodesList *[]node.Node) (node.Node, error) {

	next := make([]node.Next, 0)

	var currNode node.Node

	if n.Next != nil {
		for key, value := range n.Next {
			Node, err := getNext(key, value, registry, NodesList)
			if err != nil {
				return nil, err
			}
			next = append(next, Node)
		}
	}

	// check if Processor(and which types)
	processor, ok := registry.GetProcessor(name)
	if ok {
		if len(next) == 0 {
			return nil, fmt.Errorf("plugin [%s] has no next(s) of type output, a pipeline path must end with an output plugin", name)
		}
		switch p := processor.ProcessorBase.(type) {
		case component.ProcessorReadOnly:
			currNode = &node.ReadOnly{
				ProcessorReadOnly: p,
				Next:              next,
				Resource:          processor.Resource,
			}
		case component.ProcessorReadWrite:
			currNode = &node.ReadWrite{
				ProcessorReadWrite: p,
				Next:               next,
				Resource:           processor.Resource,
			}
		default:
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
		// Not Processor, check if output.
	} else {
		output, ok := registry.GetOutput(name)
		if ok {
			if len(next) > 0 {
				return nil, fmt.Errorf("plugin [%s] has next(s), output plugins must not have next(s)", name)
			}
			currNode = &node.Output{
				Output:   output,
				Next:     next,
				Resource: output.Resource,
			}
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	*NodesList = append(*NodesList, currNode)

	return currNode, nil
}

func getNext(name string, n config.Node, registry registery.Registry, NodesList *[]node.Node) (node.Next, error) {
	Node, err := buildTree(name, n, registry, NodesList)

	if err != nil {
		return node.Next{}, err
	}

	tc := make(chan transaction.Transaction)

	// give node its receive chan
	Node.SetTransactionChan(tc)

	return node.Next{
		Node:            Node,
		TransactionChan: tc,
	}, nil
}
