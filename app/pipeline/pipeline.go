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
)

//Pipeline Holds the recursive tree of Nodes and their next nodes, etc
type Pipeline struct {
	RecieveChan chan transaction.Transaction
	Resource    resource.Resource
	Next        node.Node
	NodesList   []*node.Node
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
			go func(txn transaction.Transaction) {
				// TODO handle context error
				responseChan := make(chan transaction.Response)
				p.Next.GetReceiverChan() <- transaction.Transaction{
					Payload:      txn.Payload,
					ImageData:    txn.ImageData,
					ResponseChan: responseChan,
				}
				txn.ResponseChan <- <-responseChan
			}(value)
		}
		p.Logger.Infow("Stopped.")
	}()

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
