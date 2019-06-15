package config

import (
	"runtime"

	"github.com/imdario/mergo"
)

// TODO use reflect to decrease code duplication (logic is almost the same here)

// DefaultComponent used in defaults
var DefaultComponent = Component{
	Concurrency: runtime.NumCPU(),
}

//DefaultInputs used in defaults
var DefaultInput = Input{
	Component: DefaultComponent,
}

//DefaultProcessor used in defaults
var DefaultProcessor = Processor{
	Component: DefaultComponent,
}

//DefaultOutput used in defaults
var DefaultOutput = Output{
	Component: DefaultComponent,
}

//DefaultAppConfig used in defaults
var DefaultAppConfig = AppConfig{
	Logger: "prod",
}

//DefaultNode used in defaults
var DefaultNode = Node{
	Async: false,
}

//DefaultPipeline used in defaults
var DefaultPipeline = Pipeline{
	Concurrency: runtime.NumCPU(),
}

//ApplyDefault func used in defaults
func (i *Input) ApplyDefault() error {
	return mergo.Merge(i, DefaultInput)
}

//ApplyDefault func used in defaults
func (p *Processor) ApplyDefault() error {
	return mergo.Merge(p, DefaultProcessor)
}

//ApplyDefault func used in defaults
func (o *Output) ApplyDefault() error {
	return mergo.Merge(o, DefaultOutput)
}

//ApplyDefault func used in defaults
func (p *Pipeline) ApplyDefault() error {
	err := mergo.Merge(p, DefaultPipeline)
	if err != nil {
		return err
	}

	for _, value := range p.Pipeline {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}

	return nil
}

//ApplyDefault func used in defaults
func (n *Node) ApplyDefault() error {
	err := mergo.Merge(n, DefaultNode)
	if err != nil {
		return err
	}
	for _, value := range n.Next {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

//ApplyDefault func used in defaults
func (n *AppConfig) ApplyDefault() error {
	return mergo.Merge(n, DefaultAppConfig)
}

//ApplyDefault func used in defaults
func (i *InputsConfig) ApplyDefault() error {
	for _, value := range i.Inputs {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

//ApplyDefault func used in defaults
func (i *OutputsConfig) ApplyDefault() error {
	for _, value := range i.Outputs {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

//ApplyDefault func used in defaults
func (i *ProcessorsConfig) ApplyDefault() error {
	for _, value := range i.Processors {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

//ApplyDefault func used in defaults
func (i *PipelinesConfig) ApplyDefault() error {
	for _, value := range i.Pipelines {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}
