package pipeline

import (
	"fmt"
	"strconv"

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

	jobChan := make(chan job.Job)

	//TODO hash pipelines

	// Create pipeline
	p := &Pipeline{
		name:           name,
		hash:           "TODOHASHPIPELINE",
		receiveJobChan: jobChan,
		Root:           nil,
		persistence:    persistence.Persistence{},
		NodeMap:        make(map[string]*node.Node),
		Logger:         *logger.Named(name),
	}

	// create persistence
	persistence, err := persistence.NewPersistence(name, "TODOHASHPIPELINE", p.Logger)
	if err != nil {
		return &wrapper{}, err
	}
	p.persistence = persistence

	// Lookup Nexts of this Node
	nexts, err := p.getNodeNexts(Config.Pipeline, registry, false)
	if err != nil {
		return &wrapper{}, err
	}

	// set pipeline root node
	// Node Beginning Dummy Node
	rootJobChan := make(chan job.Job)
	p.Root = &node.Next{
		Node:    node.NewDummy("dummy", component.NewResource(Config.Concurrency), false, nexts, p.createAsyncJob, rootJobChan, p.Logger),
		JobChan: rootJobChan,
	}

	// save root node.
	p.NodeMap[p.Root.Name] = p.Root.Node

	return &wrapper{
		Pipeline: p,
		jobChan:  jobChan,
	}, nil
}

func (p *Pipeline) getNodeNexts(next map[string]*config.Node, registry component.Registry, forceSync bool) ([]node.Next, error) {
	nexts := make([]node.Next, 0)

	for name, n := range next {

		//
		async, forceSync := evaluateAsync(n.Async, forceSync)

		// add nodeNexts
		nodeNexts, err := p.getNodeNexts(n.Next, registry, forceSync)
		if err != nil {
			return nil, err
		}

		jobChan := make(chan job.Job)

		// create node of the configure components
		currNode, err := p.createNode(name, p.getUniqueNodeName(name), async, registry, nodeNexts, jobChan, len(n.Next))
		if err != nil {
			return nil, err
		}

		// append to nodeNexts
		nexts = append(nexts, node.Next{
			Node:    currNode,
			JobChan: jobChan,
		})
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

func (p Pipeline) createNode(componentName, nodeName string, async bool, registry component.Registry,
	nexts []node.Next, jobChan chan job.Job, nextsCount int) (*node.Node, error) {

	var Node *node.Node

	Component := registry.Component(componentName)
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
			Node = node.NewReadWrite(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.Logger)
		case *component.ProcessorReadOnly:
			Node = node.NewReadOnly(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.Logger)
		case *component.ProcessorReadWriteStream:
			Node = node.NewReadWriteStream(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.Logger)
		}

	case *component.Output:
		if nextsCount > 0 {
			return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", nodeName)
		}
		Node = node.NewOutput(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.Logger)
	case *component.Input:
		return nil, fmt.Errorf("plugin [%s] is an input plugin", nodeName)
	default:
		return nil, fmt.Errorf("plugin [%s] doesn't exists", nodeName)
	}

	// Assert check.
	if Node == nil {
		return nil, fmt.Errorf("failed to create Node [%s]", componentName)
	}

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
