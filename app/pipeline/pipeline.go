package pipeline

import (
	"context"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/manager"
	"github.com/sherifabdlnaby/prism/pkg/component"
	"github.com/sherifabdlnaby/semaphore"
)

type Pipeline struct {
	RecieveChan chan component.Transaction
	Next        NodeInterface
	Sema        semaphore.Weighted
	NodesList   []*NodeInterface
}

func (p *Pipeline) Start() {
	go func() {
		for value := range p.RecieveChan {
			go func(transaction component.Transaction) {
				// TODO handle context error
				_ = p.Sema.Acquire(context.TODO(), 1)
				responseChan := make(chan component.Response)
				p.Next.GetRecieverChan() <- component.Transaction{
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

func NewPipeline(pc config.Pipeline) *Pipeline {

	next := make([]NextNode, 0)
	NodesList := make([]*NodeInterface, 0)

	beginNode := dummyNode{
		Node: Node{
			RecieverChan: make(chan component.Transaction),
		},
	}

	var node NodeInterface = &beginNode
	NodesList = append(NodesList, &node)

	for key, value := range pc.Pipeline {
		next = append(next, NextNode{
			async: value.Async,
			Node:  buildTree(key, value, &NodesList),
		})
	}

	beginNode.Next = next
	beginNode.Start()

	pip := Pipeline{
		RecieveChan: make(chan component.Transaction),
		Next:        &beginNode,
		Sema:        *semaphore.NewWeighted(int64(pc.Concurrency)),
		NodesList:   NodesList,
	}

	return &pip
}

func buildTree(name string, n config.Node, NodesList *[]*NodeInterface) NodeInterface {

	next := make([]NextNode, 0)

	var node NodeInterface = nil

	*NodesList = append(*NodesList, &node)

	if n.Next != nil {
		for key, value := range n.Next {
			next = append(next, NextNode{
				async: value.Async,
				Node:  buildTree(key, value, NodesList),
			})
		}
	}

	// processor plugins
	processor, ok := manager.GetProcessor(name)
	if ok {
		node = &ProcessingNode{
			Node: Node{
				RecieverChan: make(chan component.Transaction),
				Next:         next,
			},
			ProcessorWrapper: processor,
		}
	} else {
		output, ok := manager.GetOutput(name)
		if ok {
			node = &OutputNode{
				Node: Node{
					RecieverChan: make(chan component.Transaction),
					Next:         next,
				},
				OutputWrapper: output,
			}
		} else {
			panic("PLUGINS DOESN'T EXIT")
			//TODO return error instead
		}
	}

	node.Start()
	return node
}
