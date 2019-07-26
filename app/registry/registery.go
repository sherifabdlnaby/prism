package registry

//Registry Contains Plugin Instances
type Registry struct {
	Inputs                   map[string]*Input
	ProcessorReadOnly        map[string]*ProcessorReadOnly
	ProcessorReadWrite       map[string]*ProcessorReadWrite
	ProcessorReadWriteStream map[string]*ProcessorReadWriteStream
	Outputs                  map[string]*Output
}

//NewRegistry Register constructor.
func NewRegistry() *Registry {
	return &Registry{
		Inputs:                   make(map[string]*Input),
		ProcessorReadOnly:        make(map[string]*ProcessorReadOnly),
		ProcessorReadWrite:       make(map[string]*ProcessorReadWrite),
		ProcessorReadWriteStream: make(map[string]*ProcessorReadWriteStream),
		Outputs:                  make(map[string]*Output),
	}
}
