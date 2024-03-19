package zenmodel

import (
	"fmt"
	"time"

	"github.com/zenmodel/zenmodel/internal/errors"
	"go.uber.org/zap/zapcore"
)

// Brainprint is short of BrainLocal Blueprint
type Brainprint struct {
	// map of all neuron
	neurons map[string]*Neuron
	// map of all link
	links map[string]*Link
	// timeout for the brain, default is no timeout
	// if set, the brain will sleep after the timeout
	timeout *time.Duration
}

type BrainState string

func NewBrainPrint() *Brainprint {
	return &Brainprint{
		neurons: make(map[string]*Neuron),
		links:   make(map[string]*Link),
	}
}

func (b *Brainprint) DeepCopy() *Brainprint {
	if b == nil {
		return nil
	}
	cp := &Brainprint{
		neurons: make(map[string]*Neuron),
		links:   make(map[string]*Link),
	}
	if b.timeout != nil {
		timeout := *b.timeout
		cp.timeout = &timeout
	}
	for id, neuron := range b.neurons {
		cp.neurons[id] = neuron.DeepCopy()
	}
	for id, link := range b.links {
		cp.links[id] = link.DeepCopy()
	}
	return cp
}

func (b *Brainprint) AddNeuron(processFn func(Brain) error) string {
	neuron := newNeuron()
	neuron.bindProcessor(&DefaultProcessor{processFn: processFn})
	b.neurons[neuron.id] = neuron
	return neuron.id
}

func (b *Brainprint) AddNeuronWithProcessor(p Processor) string {
	neuron := newNeuron()
	neuron.bindProcessor(p)
	b.neurons[neuron.id] = neuron
	return neuron.id
}

func (b *Brainprint) GetNeuron(id string) *Neuron {
	return b.neurons[id]
}

func (b *Brainprint) AddLink(fromID, toID string) (string, error) {
	// check neuron exist in brain
	if fromID == EndNeuronID {
		return "", fmt.Errorf("END neuron cannot cast to any neuron")
	}
	if toID == EndNeuronID {
		b.ensureEndNeuron()
	}
	from := b.GetNeuron(fromID)
	if from == nil {
		return "", errors.ErrNeuronNotFound(fromID)
	}
	to := b.GetNeuron(toID)
	if to == nil {
		return "", errors.ErrNeuronNotFound(toID)
	}

	link := newLink(from, to)

	if err := to.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}
	if err := from.addLinkToDefaultCastGroup(link); err != nil {
		// rollback
		to.deleteTriggerGroup(link)
		// error
		return "", errors.Wrapf(err, "add link to default cast group error")
	}
	b.links[link.id] = link

	return link.id, nil
}

func (b *Brainprint) GetLink(id string) *Link {
	return b.links[id]
}

func (b *Brainprint) AddEntryLink(toID string) (string, error) {
	to := b.GetNeuron(toID)
	if to == nil {
		return "", errors.ErrNeuronNotFound(toID)
	}

	link := newEntryLink(to)

	if err := to.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}

	b.links[link.id] = link

	return link.id, nil
}

// AddEndLink add link from specific neuron to END neuron, if END neuron not exist, create it.
func (b *Brainprint) AddEndLink(fromID string) (string, error) {
	b.ensureEndNeuron()
	end := b.GetNeuron(EndNeuronID)
	from := b.GetNeuron(fromID)
	if from == nil {
		return "", errors.ErrNeuronNotFound(fromID)
	}

	link := newLink(from, end)

	if err := end.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}
	if err := from.addLinkToDefaultCastGroup(link); err != nil {
		// rollback
		end.deleteTriggerGroup(link)
		// error
		return "", errors.Wrapf(err, "add link to default cast group error")
	}

	b.links[link.id] = link

	return link.id, nil
}

// AddLinkToCastGroup add links to specific named cast group.
// if group not exist, create the group. Groups that allow empty links.
// The specified link will remove from the default group, if it originally belonged to the default group.
func (b *Brainprint) AddLinkToCastGroup(neuronID string, groupName string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.ErrLinkNotFound(id)
		}
		links = append(links, link)
	}

	return neu.addLinkToCastGroup(groupName, links...)
}

// DeleteCastGroup ...
func (b *Brainprint) DeleteCastGroup(neuronID string, groupName string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	return neu.deleteCastGroup(groupName)
}

// AddTriggerGroup ...
func (b *Brainprint) AddTriggerGroup(neuronID string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.ErrLinkNotFound(id)
		}
		links = append(links, link)
	}

	return neu.addTriggerGroup(links...)
}

// DeleteTriggerGroup ...
func (b *Brainprint) DeleteTriggerGroup(neuronID string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.ErrLinkNotFound(id)
		}
		links = append(links, link)
	}

	neu.deleteTriggerGroup(links...)

	return nil
}

// Build will build BrainLocal
func (b *Brainprint) Build(withOpts ...Option) Brain {
	bpcp := b.DeepCopy()
	brain := NewBrainLocal(*bpcp, withOpts...)

	return brain
}

// BindCastGroupSelectFunc bind custom select function of cast group, default select default cast group.
func (b *Brainprint) BindCastGroupSelectFunc(neuronID string, selectFn func(brain Brain) string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	neu.selectFn = selectFn

	return nil
}

func (b *Brainprint) HasEndNeuron() bool {
	for nid, _ := range b.neurons {
		if nid == EndNeuronID {
			return true
		}
	}

	return false
}

func (b *Brainprint) ensureEndNeuron() {
	if b.HasEndNeuron() {
		return
	}
	neuron := newEndNeuron()
	b.neurons[neuron.id] = neuron
}

func (b *Brainprint) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	err := enc.AddArray("neurons", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, neuron := range b.neurons {
			if err := ae.AppendObject(neuron); err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	err = enc.AddArray("links", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, link := range b.links {
			if err := ae.AppendObject(link); err != nil {
				return err
			}
		}

		return nil
	}))
	if err != nil {
		return err
	}

	if b.timeout != nil {
		enc.AddDuration("timeout", *b.timeout)
	}

	return nil
}
