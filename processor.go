package zenmodel

type Processor interface {
	Process(brain BrainRuntime) error
	DeepCopy() Processor
}

type DefaultProcessor struct {
	// TODO 增加 timeout, retry
	processFn func(brain BrainRuntime) error
}

func (p *DefaultProcessor) Process(brain BrainRuntime) error {
	// TODO wrap process func, wrap timeout, retry
	return p.processFn(brain)
}

func (p *DefaultProcessor) DeepCopy() Processor {
	return &DefaultProcessor{
		processFn: p.processFn,
	}
}

type EndProcessor struct{}

func (p *EndProcessor) Process(_ BrainRuntime) error {
	return nil
}

func (p *EndProcessor) DeepCopy() Processor {
	return &EndProcessor{}
}
