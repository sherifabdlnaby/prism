package pipeline

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	RecieveChan chan component.Transaction
	Next        nodeInterface
	Sema        semaphore.Weighted
	NodesList   []*nodeInterface
}

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() {
	go func() {
		for value := range p.RecieveChan {
			go func(transaction component.Transaction) {
				// TODO handle context error
				_ = p.Sema.Acquire(context.TODO(), 1)
				responseChan := make(chan component.Response)
				p.Next.getReceiverChan() <- component.Transaction{
					InputPayload: transaction.InputPayload,
					ImageData:    transaction.ImageData,
					ResponseChan: responseChan,
				}
				transaction.ResponseChan <- <-responseChan
				p.Sema.Release(1)
			}(value)
		}
	}()
}

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(pc config.Pipeline) (*Pipeline, error) {

	next := make([]nextNode, 0)
	NodesList := make([]*nodeInterface, 0)

	beginNode := dummyNode{
		node: node{
			RecieverChan: make(chan component.Transaction),
		},
	}

	var node nodeInterface = &beginNode
	NodesList = append(NodesList, &node)

	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, value, &NodesList)

		if err != nil {
			return &Pipeline{}, err
		}

		next = append(next, nextNode{
			async: value.Async,
			node:  Node,
		})
	}

	beginNode.Next = next
	beginNode.start()

	pip := Pipeline{
		RecieveChan: make(chan component.Transaction),
		Next:        &beginNode,
		Sema:        *semaphore.NewWeighted(int64(pc.Concurrency)),
		NodesList:   NodesList,
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, NodesList *[]*nodeInterface) (nodeInterface, error) {

	next := make([]nextNode, 0)

	var currNode nodeInterface

	*NodesList = append(*NodesList, &currNode)

	if n.Next != nil {
		for key, value := range n.Next {
			Node, err := buildTree(key, value, NodesList)

			if err != nil {
				return nil, err
			}

			next = append(next, nextNode{
				async: value.Async,
				node:  Node,
			})
		}
	}

	// processor plugins
	processor, ok := manager.GetProcessor(name)
	if ok {
		switch p := processor.ProcessorBase.(type) {
		case component.ProcessorReadOnly:
			currNode = &processingReadOnlyNode{
				node: node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				ResourceManager:   processor.ResourceManager,
				ProcessorReadOnly: p,
			}
		case component.ProcessorReadWrite:
			currNode = &processingReadWriteNode{
				node: node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				ResourceManager:    processor.ResourceManager,
				ProcessorReadWrite: p,
			}
		}
	} else {
		output, ok := manager.GetOutput(name)
		if ok {
			currNode = &outputNode{
				node: node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				ResourceManager: output.ResourceManager,
				Output:          output.Output,
			}
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	currNode.start()

	return currNode, nil
}
