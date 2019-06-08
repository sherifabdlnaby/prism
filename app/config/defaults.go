package config

import (
	"runtime"

	"github.com/imdario/mergo"
)

// TODO use reflect to decrease code duplication (logic is almost the same here)

var DefaultComponent = Component{
	Concurrency: runtime.NumCPU(),
}

var DefaultInput = Input{
	Component: DefaultComponent,
}

var DefaultProcessor = Processor{
	Component: DefaultComponent,
}

var DefaultOutput = Output{
	Component: DefaultComponent,
}

var DefaultAppConfig = AppConfig{
	Logger: "prod",
}

var DefaultNode = Node{
	Async: false,
}

var DefaultPipeline = Pipeline{
	Concurrency: runtime.NumCPU(),
}

func (i *Input) ApplyDefault() error {
	return mergo.Merge(i, DefaultInput)
}

func (p *Processor) ApplyDefault() error {
	return mergo.Merge(p, DefaultProcessor)
}

func (o *Output) ApplyDefault() error {
	return mergo.Merge(o, DefaultOutput)
}

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

func (n *AppConfig) ApplyDefault() error {
	return mergo.Merge(n, DefaultAppConfig)
}

func (i *InputsConfig) ApplyDefault() error {
	for _, value := range i.Inputs {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *OutputsConfig) ApplyDefault() error {
	for _, value := range i.Outputs {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *ProcessorsConfig) ApplyDefault() error {
	for _, value := range i.Processors {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *PipelinesConfig) ApplyDefault() error {
	for _, value := range i.Pipelines {
		err := value.ApplyDefault()
		if err != nil {
			return err
		}
	}
	return nil
}
