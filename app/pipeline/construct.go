package pipeline

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"go.uber.org/zap"
)

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(name string, Config config.Pipeline, registry component.Registry,
	logger zap.SugaredLogger) (*wrapper, error) {
	var err error

	tc := make(chan job.Job)

	//TODO hash pipelines

	// Create pipeline
	p := &Pipeline{
		name:           name,
		hash:           "TODOHASHPIPELINE",
		config:         Config,
		receiveJobChan: tc,
		registry:       registry,
		Root:           nil,
		persistence:    persistence.Persistence{},
		NodeMap:        make(map[string]*node.Node),
		wg:             sync.WaitGroup{},
		Logger:         *logger.Named(name),
	}

	// Node Beginning Dummy Node
	root := node.NewNext(node.NewDummy("dummy", component.NewResource(Config.Concurrency), p.Logger))

	// create persistence
	persistence, err := persistence.NewPersistence(name, "TODOHASHPIPELINE", p.Logger)
	if err != nil {
		return &wrapper{}, err
	}
	p.persistence = persistence

	// Lookup Nexts of this Node
	nexts, err := p.getNodeNexts(Config.Pipeline, false)
	if err != nil {
		return &wrapper{}, err
	}

	// set begin Node to nexts (Pipelines beginning)
	root.SetNexts(nexts)

	// set pipeline root node
	p.Root = root

	// save root node.
	p.NodeMap[root.Name] = root.Node

	return &wrapper{
		Pipeline: p,
		jobChan:  tc,
	}, nil
}

func (p *Pipeline) getNodeNexts(next map[string]*config.Node, forceSync bool) ([]node.Next, error) {
	nexts := make([]node.Next, 0)

	for name, n := range next {

		//
		async, forceSync := evaluateAsync(n.Async, forceSync)

		// create node of the configure components
		currNode, err := p.createNode(name, p.getUniqueNodeName(name), async, p.persistence, len(n.Next))
		if err != nil {
			return nil, err
		}

		// add nodeNexts
		nodeNexts, err := p.getNodeNexts(n.Next, forceSync)
		if err != nil {
			return nil, err
		}

		// set nodeNexts
		currNode.SetNexts(nodeNexts)

		// create a next wrapper
		next := node.NewNext(currNode)

		// append to nodeNexts
		nexts = append(nexts, *next)
	}

	return nexts, nil
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

func (p Pipeline) createNode(componentName, nodeName string, async bool, persistence persistence.Persistence,
	nextsCount int) (*node.Node, error) {
	var Node *node.Node

	// check if ProcessReadWrite(and which types)
	Component := p.registry.Component(componentName)
	if Component == nil {
		return nil, fmt.Errorf("plugin [%s] doesn't exists", componentName)
	}

	switch Component := Component.(type) {
	case *component.ProcessorReadWrite, *component.ProcessorReadOnly, *component.ProcessorReadWriteStream:
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", nodeName)
		}

		switch Component := Component.(type) {
		case *component.ProcessorReadWrite:
			Node = node.NewReadWrite(nodeName, Component, p.Logger)
		case *component.ProcessorReadOnly:
			Node = node.NewReadOnly(nodeName, Component, p.Logger)
		case *component.ProcessorReadWriteStream:
			Node = node.NewReadWriteStream(nodeName, Component, p.Logger)
		}

	case *component.Output:
		if nextsCount > 0 {
			return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", nodeName)
		}
		Node = node.NewOutput(nodeName, Component, p.Logger)
	case *component.Input:
		return nil, fmt.Errorf("plugin [%s] is an input plugin", nodeName)
	default:
		return nil, fmt.Errorf("plugin [%s] doesn't exists", nodeName)
	}

	Node.SetAsync(async)

	Node.SetPersistence(p.persistence)

	// save in map
	p.NodeMap[nodeName] = Node

	return Node, nil
}

func evaluateAsync(async, forceSync bool) (bool, bool) {
	if forceSync {
		async = false
	}
	// all NEXT nodes to be sync if current is async.
	if async {
		forceSync = true
	}
	return async, forceSync
}
