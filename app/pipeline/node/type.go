package node

import (
	"sync"

	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"go.uber.org/zap"
)

type Type interface {
	job(t job.Job)
	jobStream(t job.Job)
}

func newBase(name string, nodeType Type, resource *component.Resource, logger zap.SugaredLogger) *Node {
	return &Node{
		Name:     name,
		async:    false,
		wg:       sync.WaitGroup{},
		nexts:    nil,
		resource: resource,
		nodeType: nodeType,
		Logger:   *logger.Named(name),
	}
}

//NewReadOnly Construct a new ReadOnly node
func NewReadOnly(name string, processor *component.ProcessorReadOnly, logger zap.SugaredLogger) *Node {
	nodeType := &readOnly{processor: processor}
	base := newBase(name, nodeType, &processor.Resource, logger)
	nodeType.Node = base
	return nodeType.Node
}

//NewReadWrite Construct a new ReadWrite Node
func NewReadWrite(name string, processor *component.ProcessorReadWrite, logger zap.SugaredLogger) *Node {
	nodeType := &readWrite{processor: processor}
	base := newBase(name, nodeType, &processor.Resource, logger)
	nodeType.Node = base
	return nodeType.Node
}

//NewReadWriteStream Construct a new ReadWriteStream Node
func NewReadWriteStream(name string, processor *component.ProcessorReadWriteStream, logger zap.SugaredLogger) *Node {
	nodeType := &readWriteStream{processor: processor}
	base := newBase(name, nodeType, &processor.Resource, logger)
	nodeType.Node = base
	return nodeType.Node
}

//NewOutput Construct a new Output Node
func NewOutput(name string, out *component.Output, logger zap.SugaredLogger) *Node {
	nodeType := &output{output: out}
	base := newBase(name, nodeType, &out.Resource, logger)
	nodeType.Node = base
	return nodeType.Node
}

//NewDummy Construct a new Dummy Node
func NewDummy(name string, r *component.Resource, logger zap.SugaredLogger) *Node {
	nodeType := &dummy{}
	base := newBase(name, nodeType, r, logger)
	nodeType.Node = base
	return nodeType.Node
}
