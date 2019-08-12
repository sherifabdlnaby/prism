package pipeline

import (
	"fmt"
	"strconv"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/app/config"
	"github.com/sherifabdlnaby/prism/app/pipeline/node"
	"github.com/sherifabdlnaby/prism/app/pipeline/persistence"
	"github.com/sherifabdlnaby/prism/pkg/job"
)

//NewPipeline Construct a NewPipeline using config.
func (m *Manager) NewPipeline(name string, Config config.Pipeline) (*wrapper, error) {
	var err error

	jobChan := make(chan job.Job)

	//TODO hash pipelines

	// Create pipeline
	p := &pipeline{
		name:           name,
		hash:           "TODOHASHPIPELINE",
		root:           nil,
		resource:       *component.NewResource(Config.Concurrency),
		receiveJobChan: jobChan,
		nodeMap:        make(map[string]*node.Node),
		bucket:         persistence.Bucket{},
		logger:         *m.logger.Named(name),
	}

	// create bucket
	persistence, err := m.persistence.Bucket(name, "TODOHASHPIPELINE", p.logger)
	if err != nil {
		return &wrapper{}, err
	}
	p.bucket = *persistence

	// Lookup Nexts of this Node
	nexts, err := p.getNodeNexts(Config.Pipeline, m.registry, false)
	if err != nil {
		return &wrapper{}, err
	}

	// set pipeline root node
	// Node Beginning Dummy Node
	rootJobChan := make(chan job.Job)
	p.root = &node.Next{
		Node:    node.NewDummy(nexts, rootJobChan, p.logger),
		JobChan: rootJobChan,
	}

	return &wrapper{
		pipeline: p,
		jobChan:  jobChan,
	}, nil
}

func (p *pipeline) getNodeNexts(next map[string]*config.Node, registry component.Registry, forceSync bool) ([]node.Next, error) {
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

func (p *pipeline) getUniqueNodeName(name string) string {
	for i := 0; ; {
		_, ok := p.nodeMap[name]
		if !ok {
			return name
		}
		name += "_" + strconv.Itoa(i)
	}
}

func (p *pipeline) createNode(componentName, nodeName string, async bool, registry component.Registry,
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
			Node = node.NewReadWrite(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.logger)
		case *component.ProcessorReadOnly:
			Node = node.NewReadOnly(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.logger)
		case *component.ProcessorReadWriteStream:
			Node = node.NewReadWriteStream(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.logger)
		}

	case *component.Output:
		if nextsCount > 0 {
			return nil, fmt.Errorf("plugin [%s] has nexts(s), output plugins must not have nexts(s)", nodeName)
		}
		Node = node.NewOutput(nodeName, Component, async, nexts, p.createAsyncJob, jobChan, p.logger)
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
	p.nodeMap[nodeName] = Node

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
