package zenmodel

import (
	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/core"
	"github.com/zenmodel/zenmodel/internal/errors"
	"github.com/zenmodel/zenmodel/internal/utils"
	"github.com/zenmodel/zenmodel/processor"
	"github.com/zenmodel/zenmodel/pyprocessor"
)

// NewBlueprint new blueprint
func NewBlueprint() core.Blueprint {
	return &brainprint{
		id:      utils.GenID(),
		labels:  make(map[string]string),
		neurons: make(map[string]*neuron),
		links:   make(map[string]*link),
	}
}

// NewMultiLangBlueprint new multi-language blueprint
func NewMultiLangBlueprint() core.MultiLangBlueprint {
	return &brainprint{
		id:      utils.GenID(),
		labels:  make(map[string]string),
		neurons: make(map[string]*neuron),
		links:   make(map[string]*link),
	}
}

// brainprint is implement Brainprint
type brainprint struct {
	// 标识，拷贝的brainprint ID 也相同
	id string
	// labels
	labels map[string]string
	// map of all neuron
	neurons map[string]*neuron
	// map of all link
	links map[string]*link
}

func (b *brainprint) GetID() string {
	return b.id
}

func (b *brainprint) GetLabels() map[string]string {
	return b.labels
}

func (b *brainprint) SetLabels(labels map[string]string) {
	b.labels = labels
}

func (b *brainprint) GetNeuron(neuronID string) (core.Neuron, error) {
	n, ok := b.neurons[neuronID]
	if !ok {
		return nil, errors.ErrNeuronNotFound(neuronID)
	}
	return n, nil
}

func (b *brainprint) HasNeuron(neuronID string) bool {
	_, ok := b.neurons[neuronID]
	return ok
}

func (b *brainprint) ListNeurons() []core.Neuron {
	if len(b.neurons) == 0 {
		return nil
	}
	neurons := make([]core.Neuron, 0, len(b.neurons))
	for _, n := range b.neurons {
		neurons = append(neurons, n)
	}
	return neurons
}

func (b *brainprint) GetSrcNeuron(linkID string) (core.Neuron, error) {
	l, err := b.GetLink(linkID)
	if err != nil {
		return nil, err
	}

	return b.GetNeuron(l.GetSrcNeuronID())
}

func (b *brainprint) GetDestNeuron(linkID string) (core.Neuron, error) {
	l, err := b.GetLink(linkID)
	if err != nil {
		return nil, err
	}

	return b.GetNeuron(l.GetDestNeuronID())
}

func (b *brainprint) GetLink(linkID string) (core.Link, error) {
	l, ok := b.links[linkID]
	if !ok {
		return nil, errors.ErrLinkNotFound(linkID)
	}
	return l, nil
}

func (b *brainprint) HasLink(linkID string) bool {
	_, ok := b.links[linkID]
	return ok
}

func (b *brainprint) HasEntryLink() bool {
	for _, l := range b.links {
		if l.IsEntryLink() {
			return true
		}
	}

	return false
}

func (b *brainprint) HasEndLink() bool {
	for _, l := range b.links {
		if l.IsEndLink() {
			return true
		}
	}

	return false
}

func (b *brainprint) ListLinks() []core.Link {
	ret := make([]core.Link, 0, len(b.links))
	for _, l := range b.links {
		ret = append(ret, l)
	}

	return ret
}

func (b *brainprint) ListEntryLinks() []core.Link {
	ret := make([]core.Link, 0)
	for _, l := range b.links {
		if l.IsEntryLink() {
			ret = append(ret, l)
		}
	}

	return ret
}

func (b *brainprint) ListEndLinks() []core.Link {
	ret := make([]core.Link, 0)
	for _, l := range b.links {
		if l.IsEndLink() {
			ret = append(ret, l)
		}
	}

	return ret
}

func (b *brainprint) ListInLinks(neuronID string) []core.Link {
	if !b.HasNeuron(neuronID) {
		return nil
	}
	ret := make([]core.Link, 0, len(b.links))
	for _, l := range b.links {
		if l.dest == neuronID {
			ret = append(ret, l)
		}
	}

	return ret
}

func (b *brainprint) ListOutLinks(neuronID string) []core.Link {
	if !b.HasNeuron(neuronID) {
		return nil
	}
	ret := make([]core.Link, 0, len(b.links))
	for _, l := range b.links {
		if l.src == neuronID {
			ret = append(ret, l)
		}
	}

	return ret
}

func (b *brainprint) AddNeuron(processFn func(bc processor.BrainContext) error, withOpts ...core.NeuronOption) core.Neuron {
	return b.addNeuronWithProcessor(processor.NewFuncProcessor(processFn), withOpts...)
}

func (b *brainprint) AddNeuronWithProcessor(processor processor.Processor, withOpts ...core.NeuronOption) core.Neuron {
	return b.addNeuronWithProcessor(processor, withOpts...)
}

func (b *brainprint) AddNeuronWithPyProcessor(pyCodePath, moduleName, processorClassName string, withOpts ...core.NeuronOption) core.Neuron {
	processor := pyprocessor.LoadPythonProcessor(pyCodePath, moduleName, processorClassName)
	return b.addNeuronWithProcessor(processor, withOpts...)
}

func (b *brainprint) AddLink(from, to core.Neuron, withOpts ...core.LinkOption) (core.Link, error) {
	// validate
	src, ok := b.neurons[from.GetID()]
	if !ok {
		return nil, errors.ErrNeuronNotFound(from.GetID())
	}
	dest, ok := b.neurons[to.GetID()]
	if !ok {
		return nil, errors.ErrNeuronNotFound(to.GetID())
	}
	// new link, and neurons set
	l := newLink(from.GetID(), to.GetID())
	src.addOutLink(l.GetID())
	dest.addInLink(l.GetID())

	// bp add link
	for _, opt := range withOpts {
		opt.Apply(l)
	}
	b.links[l.GetID()] = l

	return l, nil
}

func (b *brainprint) AddEntryLinkTo(to core.Neuron, withOpts ...core.LinkOption) (core.Link, error) {
	// validate
	dest, ok := b.neurons[to.GetID()]
	if !ok {
		return nil, errors.ErrNeuronNotFound(to.GetID())
	}
	// new link, and neurons set
	l := newEntryLink(to.GetID())
	dest.addInLink(l.GetID())

	// bp add link
	for _, opt := range withOpts {
		opt.Apply(l)
	}
	b.links[l.GetID()] = l

	return l, nil
}

func (b *brainprint) AddEndLinkFrom(from core.Neuron, withOpts ...core.LinkOption) (core.Link, error) {
	// validate
	src, ok := b.neurons[from.GetID()]
	if !ok {
		return nil, errors.ErrNeuronNotFound(from.GetID())
	}
	// ensure END neuron
	end := b.ensureEndNeuron()
	// new link, and neurons set
	l := newEndLink(src.GetID())
	src.addOutLink(l.GetID())
	end.addInLink(l.GetID())

	// bp add link
	for _, opt := range withOpts {
		opt.Apply(l)
	}
	b.links[l.GetID()] = l

	return l, nil
}

func (b *brainprint) Clone() core.Blueprint {
	if b == nil {
		return nil
	}
	cp := &brainprint{
		id:      b.id,
		labels:  utils.LabelsDeepCopy(b.labels),
		neurons: make(map[string]*neuron),
		links:   make(map[string]*link),
	}
	for id, n := range b.neurons {
		cp.neurons[id] = n.deepCopy()
	}
	for id, l := range b.links {
		cp.links[id] = l.deepCopy()
	}
	return cp
}

func (b *brainprint) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", b.id).
		Any("labels", b.labels).
		Array("links", linkArray(b.links)).
		Array("neurons", neuArray(b.neurons))
}

type neuArray map[string]*neuron
type linkArray map[string]*link

func (ns neuArray) MarshalZerologArray(a *zerolog.Array) {
	for _, n := range ns {
		a.Object(n)
	}
}

func (ls linkArray) MarshalZerologArray(a *zerolog.Array) {
	for _, l := range ls {
		a.Object(l)
	}
}

func (b *brainprint) addNeuronWithProcessor(p processor.Processor, withOpts ...core.NeuronOption) core.Neuron {
	n := newNeuron(p)
	for _, opt := range withOpts {
		opt.Apply(n)
	}
	b.neurons[n.GetID()] = n

	return n
}

func (b *brainprint) ensureEndNeuron() *neuron {
	n, ok := b.neurons[core.EndNeuronID]
	if ok {
		return n
	}

	n = newEndNeuron()
	b.neurons[n.GetID()] = n

	return n
}
