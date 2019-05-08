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
	ReceiveChan chan transaction.Transaction
	Resource    resource.Resource
	Next        node.Node
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

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	if p.status != new {
		// set status = started (no need atomic here, just for sake of consistency)
		atomic.SwapInt32((*int32)(&p.status), int32(started))

		return nil
	}

	// set status = started (no need atomic here, just for sake of consistency)
	atomic.SwapInt32((*int32)(&p.status), int32(started))

	go func() {
		for value := range p.ReceiveChan {
			if p.status != started {
				value.ResponseChan <- response.Error(fmt.Errorf("pipeline is not started, request terminated"))
				continue
			}
			p.wg.Add(1)
			go p.job(value)
		}
	}()

	return nil
}

//Stop stops the pipeline, that means that any transaction received on this pipeline after stopping will return
// error response unless re-started again.
func (p *Pipeline) Stop() error {
	atomic.SwapInt32((*int32)(&p.status), int32(closed))
	p.wg.Wait()
	return nil
}

func (p *Pipeline) job(txn transaction.Transaction) {
	responseChan := make(chan response.Response, 1)
	p.Next.GetReceiverChan() <- transaction.Transaction{
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

	next := make([]node.Next, 0)
	NodesList := make([]node.Node, 0)
	resource := resource.NewResource(pc.Concurrency, logger)

	beginNode := node.Dummy{
		RecieverChan: make(chan transaction.Transaction),
		Resource:     *resource,
	}

	NodesList = append(NodesList, &beginNode)

	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, value, registry, &NodesList)

		if err != nil {
			resource.Logger.Error(err.Error())
			return &Pipeline{}, err
		}

		next = append(next, node.Next{
			Async: value.Async,
			Node:  Node,
		})
	}

	beginNode.Next = next

	pip := Pipeline{
		ReceiveChan: make(chan transaction.Transaction),
		Next:        &beginNode,
		NodesList:   NodesList,
		Logger:      logger,
		Resource:    *resource,
		status:      new,
	}

	// start pipeline node wrappers
	for _, value := range pip.NodesList {
		value.Start()
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registery.Registry, NodesList *[]node.Node) (node.Node, error) {

	next := make([]node.Next, 0)

	var currNode node.Node

	if n.Next != nil {
		for key, value := range n.Next {
			Node, err := buildTree(key, value, registry, NodesList)

			if err != nil {
				return Node, err
			}

			next = append(next, node.Next{
				Async: value.Async,
				Node:  Node,
			})
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
				ReceiverChan:      make(chan transaction.Transaction),
				Next:              next,
				Resource:          processor.Resource,
			}
		case component.ProcessorReadWrite:
			currNode = &node.ReadWrite{
				ProcessorReadWrite: p,
				ReceiverChan:       make(chan transaction.Transaction),
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
				Output:       output,
				ReceiverChan: make(chan transaction.Transaction),
				Next:         next,
				Resource:     output.Resource,
			}
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	*NodesList = append(*NodesList, currNode)

	currNode.Start()

	return currNode, nil
}
