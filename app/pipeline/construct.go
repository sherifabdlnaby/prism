package pipeline

import (
	"fmt"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/app/registry"
	"github.com/sherifabdlnaby/prism/app/registry/wrapper"
	"github.com/sherifabdlnaby/prism/app/resource"
	"github.com/sherifabdlnaby/prism/pkg/transaction"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

//NewPipeline Construct a NewPipeline using config.
func NewPipeline(name string, Config config.Pipeline, registry registry.Registry, tc <-chan transaction.Transaction,
	logger zap.SugaredLogger, hash string) (*Pipeline, error) {
	var err error

	// Create pipeline
	p := &Pipeline{
		name:           name,
		hash:           hash,
		config:         Config,
		receiveTxnChan: tc,
		registry:       registry,
		Root:           nil,
		persistence:    persistence.Persistence{},
		NodeMap:        make(map[string]*node.Node),
		wg:             sync.WaitGroup{},
		Logger:         *logger.Named(name),
	}

	// Node Beginning Dummy Node
	root := node.NewNext(node.NewDummy("dummy", resource.NewResource(Config.Concurrency), p.Logger))

	// create persistence
	persistence, err := persistence.NewPersistence(name, hash, p.Logger)
	if err != nil {
		return &Pipeline{}, err
	}
	p.persistence = persistence

	// Lookup Nexts of this Node
	nexts, err := p.getNodeNexts(Config.Pipeline, false)
	if err != nil {
		return &Pipeline{}, err
	}

	// set begin Node to nexts (Pipeline beginning)
	root.SetNexts(nexts)

	// set pipeline root node
	p.Root = root

	// save root node.
	p.NodeMap[root.Name] = root.Node

	return p, nil
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
	component := p.registry.GetComponent(componentName)
	if component == nil {
		return nil, fmt.Errorf("plugin [%s] doesn't exists", componentName)
	}

	switch component := component.(type) {
	case *wrapper.ProcessorReadWrite, *wrapper.ProcessorReadOnly, *wrapper.ProcessorReadWriteStream:
		if nextsCount == 0 {
			return nil, fmt.Errorf("plugin [%s] has no nexts(s) of type output, a pipeline path must end with an output plugin", nodeName)
		}

		switch component := component.(type) {
		case *wrapper.ProcessorReadWrite:
			Node = node.NewReadWrite(nodeName, component, p.Logger)
		case *wrapper.ProcessorReadOnly:
			Node = node.NewReadOnly(nodeName, component, p.Logger)
		case *wrapper.ProcessorReadWriteStream:
			Node = node.NewReadWriteStream(nodeName, component, p.Logger)
		}

	case *wrapper.Output:
		if nextsCount > 0 {
			return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", nodeName)
		}
		Node = node.NewOutput(nodeName, component, p.Logger)
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
