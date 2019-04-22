package pipeline

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/registery"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
	"go.uber.org/zap"
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	RecieveChan chan component.Transaction
	Next        Interface
	Sema        semaphore.Weighted
	NodesList   []*Interface
	Logger      zap.SugaredLogger
}

//Start starts the pipeline and start accepting Input
func (p *Pipeline) Start() error {
	// start pipeline node wrappers
	for _, value := range p.NodesList {
		(*value).Start()
	}

	go func() {
		for value := range p.RecieveChan {
			go func(transaction component.Transaction) {
				// TODO handle context error
				_ = p.Sema.Acquire(context.TODO(), 1)
				responseChan := make(chan component.Response)
				p.Next.GetReceiverChan() <- component.Transaction{
					InputPayload: transaction.InputPayload,
					ImageData:    transaction.ImageData,
					ResponseChan: responseChan,
				}
				transaction.ResponseChan <- <-responseChan
				p.Sema.Release(1)
			}(value)
		}
		p.Logger.Infow("Stopped.")
	}()

	return nil
}

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(pc config.Pipeline, registry registery.Registry, logger zap.SugaredLogger) (*Pipeline, error) {

	next := make([]NextNode, 0)
	NodesList := make([]*Interface, 0)

	beginNode := DummyNode{
		Node: Node{
			RecieverChan: make(chan component.Transaction),
		},
	}

	var currNode Interface = &beginNode
	NodesList = append(NodesList, &currNode)

	for key, value := range pc.Pipeline {
		Node, err := buildTree(key, value, registry, &NodesList)

		if err != nil {
			return &Pipeline{}, err
		}

		next = append(next, NextNode{
			Async: value.Async,
			Node:  Node,
		})
	}

	beginNode.Next = next

	pip := Pipeline{
		RecieveChan: make(chan component.Transaction),
		Next:        &beginNode,
		Sema:        *semaphore.NewWeighted(int64(pc.Concurrency)),
		NodesList:   NodesList,
		Logger:      logger,
	}

	return &pip, nil
}

func buildTree(name string, n config.Node, registry registery.Registry, NodesList *[]*Interface) (Interface, error) {

	next := make([]NextNode, 0)

	var currNode Interface

	*NodesList = append(*NodesList, &currNode)

	if n.Next != nil {
		for key, value := range n.Next {
			Node, err := buildTree(key, value, registry, NodesList)

			if err != nil {
				return nil, err
			}

			next = append(next, NextNode{
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
			currNode = &ProcessingReadOnlyNode{
				Node: Node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				Resource:          processor.Resource,
				ProcessorReadOnly: p,
			}
		case component.ProcessorReadWrite:
			currNode = &ProcessingReadWriteNode{
				Node: Node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				Resource:           processor.Resource,
				ProcessorReadWrite: p,
			}
		}
	} else {
		output, ok := registry.GetOutput(name)
		if ok {
			currNode = &OutputNode{
				Node: Node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				Resource: output.Resource,
				Output:   output.Output,
			}
		} else {
			return nil, fmt.Errorf("plugin [%s] doesn't exists", name)
		}
	}

	currNode.Start()

	return currNode, nil
}
