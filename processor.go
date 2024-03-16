package zenmodel

type Processor interface {
	Process(brain Brain) error
	DeepCopy() Processor
}

type DefaultProcessor struct {
	// TODO 增加 timeout, retry
	processFn func(brain Brain) error
}

func (p *DefaultProcessor) Process(brain Brain) error {
	// TODO wrap process func, wrap timeout, retry
	return p.processFn(brain)
}

func (p *DefaultProcessor) DeepCopy() Processor {
	return &DefaultProcessor{
		processFn: p.processFn,
	}
}

type EndProcessor struct{}

func (p *EndProcessor) Process(brain Brain) error {
	return nil
}

func (p *EndProcessor) DeepCopy() Processor {
	return &EndProcessor{}
}
