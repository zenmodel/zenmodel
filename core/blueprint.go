package core

import "github.com/zenmodel/zenmodel/processor"

type Blueprint interface {
	GetID() string
	GetLabels() map[string]string
	SetLabels(labels map[string]string)

	GetNeuron(neuronID string) (Neuron, error)
	HasNeuron(neuronID string) bool
	ListNeurons() []Neuron
	GetSrcNeuron(linkID string) (Neuron, error)
	GetDestNeuron(linkID string) (Neuron, error)

	GetLink(linkID string) (Link, error)
	HasLink(linkID string) bool
	HasEntryLink() bool
	HasEndLink() bool
	ListLinks() []Link
	ListEntryLinks() []Link
	ListEndLinks() []Link
	ListInLinks(neuronID string) []Link
	ListOutLinks(neuronID string) []Link

	AddNeuron(processFn func(bc processor.BrainContext) error, withOpts ...NeuronOption) Neuron
	AddNeuronWithProcessor(processor processor.Processor, withOpts ...NeuronOption) Neuron
	AddLink(from, to Neuron, withOpts ...LinkOption) (Link, error)
	AddEntryLinkTo(neuron Neuron, withOpts ...LinkOption) (Link, error)
	AddEndLinkFrom(neuron Neuron, withOpts ...LinkOption) (Link, error)

	Clone() Blueprint
}

// MultiLangBlueprint is extension interface of Blueprint, it is used for supporting multi-language blueprint
type MultiLangBlueprint interface {
	Blueprint
	// AddNeuronWithPyProcessor add a neuron with a python processor
	AddNeuronWithPyProcessor(pyCodePath, moduleName, processorClassName string, withOpts ...NeuronOption) Neuron
}
