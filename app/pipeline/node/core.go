package node

import (
	"github.com/sherifabdlnaby/prism/app/component"
	"github.com/sherifabdlnaby/prism/pkg/job"
	"go.uber.org/zap"
)

type core interface {
	process(j job.Job)
	processStream(j job.Job)
}

func newBase(name string, core core, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, resource *component.Resource,
	logger zap.SugaredLogger) *Node {
	return &Node{
		Name:           name,
		async:          async,
		nexts:          nexts,
		core:           core,
		createAsyncJob: createAsync,
		resource:       resource,
		logger:         *logger.Named(name),
		receiveJobChan: jobChan,
	}
}

//NewReadOnly Construct a new ReadOnly node
func NewReadOnly(name string, processor *component.ProcessorReadOnly, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readOnly{processor: processor}
	base := newBase(name, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewReadWrite Construct a new ReadWrite Node
func NewReadWrite(name string, processor *component.ProcessorReadWrite, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readWrite{processor: processor}
	base := newBase(name, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewReadWriteStream Construct a new ReadWriteStream Node
func NewReadWriteStream(name string, processor *component.ProcessorReadWriteStream, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readWriteStream{processor: processor}
	base := newBase(name, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewOutput Construct a new Output Node
func NewOutput(name string, out *component.Output, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &output{output: out}
	base := newBase(name, core, async, nexts, createAsync, jobChan, &out.Resource, logger)
	core.Node = base
	return core.Node
}

//NewDummy Construct a new Dummy Node
func NewDummy(name string, r *component.Resource, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &dummy{}
	base := newBase(name, core, async, nexts, createAsync, jobChan, r, logger)
	core.Node = base
	return core.Node
}
