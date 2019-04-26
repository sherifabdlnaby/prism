package pipeline

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

type (
	//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
	Pipeline struct {
		RecieveChan chan transaction.Transaction
		Resource    resource.Resource
		Next        node.Node
		NodesList   []*node.Node
		Logger      zap.SugaredLogger
		wg          sync.WaitGroup
		status      status
	}
	status int32
)

const (
	_              = iota // ignore first value by assigning to blank identifier
	new     status = 1 + iota
	started status = 1 + iota
	closed  status = 1 + iota
)

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {

	if p.status == new {
		// set status = started (no need atomic here, just for sake of consistency)
		atomic.SwapInt32((*int32)(&p.status), int32(started))

		go func() {
			for value := range p.RecieveChan {
				if p.status != started {
					value.ResponseChan <- transaction.ResponseError(fmt.Errorf("pipeline is not started, request terminated"))
					continue
				}
				p.wg.Add(1)
				go func(txn transaction.Transaction) {
					// TODO handle context error
					responseChan := make(chan transaction.Response)
					p.Next.GetReceiverChan() <- transaction.Transaction{
						Payload:      txn.Payload,
						ImageData:    txn.ImageData,
						ResponseChan: responseChan,
					}
					txn.ResponseChan <- <-responseChan
					p.wg.Done()
				}(value)
			}
		}()

		return nil
	}

	// set status = started (no need atomic here, just for sake of consistency)
	atomic.SwapInt32((*int32)(&p.status), int32(started))

	return nil
}

//Stop stops the pipeline, that means that any transaction received on this pipeline after stopping will return
// error response unless re-started again.
func (p *Pipeline) Stop() error {
	atomic.SwapInt32((*int32)(&p.status), int32(closed))
	p.wg.Wait()
	return nil
}

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(pc config.Pipeline, registry registery.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	next := make([]node.NextNode, 0)
	NodesList := make([]*node.Node, 0)
	resource := resource.NewResource(pc.Concurrency, logger)

	beginNode := node.Node{
		RecieverChan: make(chan transaction.Transaction),
		Resource:     *resource,
		Component:    &node.DummyNode{},
	}

	NodesList = append(NodesList, &beginNode)

	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, value, registry, &NodesList)

		if err != nil {
			resource.Logger.Error(err.Error())
			return &Pipeline{}, err
		}

		next = append(next, node.NextNode{
			Async: value.Async,
			Node:  Node,
		})
	}

	beginNode.Next = next

	pip := Pipeline{
		RecieveChan: make(chan transaction.Transaction),
		Next:        beginNode,
		NodesList:   NodesList,
		Logger:      logger,
		Resource:    *resource,
	}

	// start pipeline node wrappers
	for _, value := range pip.NodesList {
		(*value).Start()
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registery.Registry, NodesList *[]*node.Node) (node.Node, error) {

	next := make([]node.NextNode, 0)

	var currNode node.Node

	*NodesList = append(*NodesList, &currNode)

	if n.Next != nil {
		for key, value := range n.Next {
			Node, err := buildTree(key, value, registry, NodesList)

			if err != nil {
				return Node, err
			}

			next = append(next, node.NextNode{
				Async: value.Async,
				Node:  Node,
			})
		}
	}

	// processor plugins
	processor, ok := registry.GetProcessor(name)
	if ok {
		switch p := processor.ProcessorBase.(type) {
		case component.ProcessorReadOnly:
			currNode = node.Node{
				Component: &node.ReadOnly{
					ProcessorReadOnly: p,
				},
				RecieverChan: make(chan transaction.Transaction),
				Next:         next,
				Resource:     processor.Resource,
			}
		case component.ProcessorReadWrite:
			currNode = node.Node{
				Component: &node.ReadWrite{
					ProcessorReadWrite: p,
				},
				RecieverChan: make(chan transaction.Transaction),
				Next:         next,
				Resource:     processor.Resource,
			}
		}
	} else {
		output, ok := registry.GetOutput(name)
		if ok {
			currNode = node.Node{
				Component: &node.Output{
					Output: output,
				},
				RecieverChan: make(chan transaction.Transaction),
				Next:         next,
				Resource:     output.Resource,
			}
		} else {
			return node.Node{}, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	currNode.Start()

	return currNode, nil
}
