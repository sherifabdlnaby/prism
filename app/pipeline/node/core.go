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

func newBase(id ID, core core, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, resource *component.Resource,
	logger zap.SugaredLogger) *Node {
	return &Node{
		ID:             id,
		async:          async,
		nexts:          nexts,
		core:           core,
		createAsyncJob: createAsync,
		resource:       resource,
		logger:         *logger.Named(string(id)),
		receiveJobChan: jobChan,
	}
}

//NewReadOnly Construct a new ReadOnly node
func NewReadOnly(ID ID, processor *component.ProcessorReadOnly, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readOnly{processor: processor}
	base := newBase(ID, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewReadWrite Construct a new ReadWrite Node
func NewReadWrite(ID ID, processor *component.ProcessorReadWrite, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readWrite{processor: processor}
	base := newBase(ID, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewReadWriteStream Construct a new ReadWriteStream Node
func NewReadWriteStream(ID ID, processor *component.ProcessorReadWriteStream, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &readWriteStream{processor: processor}
	base := newBase(ID, core, async, nexts, createAsync, jobChan, &processor.Resource, logger)
	core.Node = base
	return core.Node
}

//NewOutput Construct a new Output Node
func NewOutput(ID ID, out *component.Output, async bool, nexts []Next,
	createAsync createAsyncFunc, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &output{output: out}
	base := newBase(ID, core, async, nexts, createAsync, jobChan, &out.Resource, logger)
	core.Node = base
	return core.Node
}

//NewDummy Construct a new Dummy Node
func NewDummy(nexts []Next, jobChan <-chan job.Job, logger zap.SugaredLogger) *Node {
	core := &dummy{}
	base := newBase("", core, false, nexts, nil, jobChan, nil, logger)
	core.Node = base
	return core.Node
}
