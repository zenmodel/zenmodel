package brain

import "github.com/zenmodel/zenmodel/processor"

const (
	EndNeuronID = "__END_NEURON__"
)

type NeuronState string

const (
	NeuronStateInactive  NeuronState = "Inactive"
	NeuronStateActivated NeuronState = "Active"
)

type Neuron interface {
	GetID() string
	GetLabels() map[string]string
	GetProcessor() processor.Processor
	GetSelector() processor.Selector
	ListInLinkIDs() []string
	ListOutLinkIDs() []string
	ListTriggerGroups() map[string][]string
	ListCastGroups() map[string][]string

	SetLabels(labels map[string]string)
	AddTriggerGroup(links ...Link) error
	AddCastGroup(groupName string, links ...Link) error
	BindCastGroupSelectFunc(selectFn func(bcr processor.BrainContextReader) string)
	BindCastGroupSelector(selector processor.Selector)
}

// NeuronOption configures a neuron.
type NeuronOption interface {
	Apply(neuron Neuron)
}

// neuronOptionFunc wraps a func, so it satisfies the NeuronOption interface.
type neuronOptionFunc func(Neuron)

func (f neuronOptionFunc) Apply(neuron Neuron) {
	f(neuron)
}

// WithNeuronLabels sets the specific labels for Neuron
func WithNeuronLabels(labels map[string]string) NeuronOption {
	return neuronOptionFunc(func(neuron Neuron) {
		neuron.SetLabels(labels)
	})
}

// WithSelectFn sets the specific selectFn for Neuron
func WithSelectFn(selectFn func(brain processor.BrainContextReader) string) NeuronOption {
	return neuronOptionFunc(func(neuron Neuron) {
		neuron.BindCastGroupSelectFunc(selectFn)
	})
}

// WithSelector sets the specific WithSelector for Neuron
func WithSelector(selector processor.Selector) NeuronOption {
	return neuronOptionFunc(func(neuron Neuron) {
		neuron.BindCastGroupSelector(selector)
	})
}
