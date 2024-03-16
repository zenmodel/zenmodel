package zenmodel

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/zenmodel/zenmodel/utils"
)

// Brainprint is short of BrainLocal Blueprint
type Brainprint struct {
	// map of all neurons
	neuronMap map[string]*Neuron
	// map of all neuron links
	linkMap map[string]*Link
	// timeout for the brain, default is no timeout
	// if set, the brain will sleep after the timeout
	timeout *time.Duration
}

type BrainState string

func NewBrainPrint() *Brainprint {
	return &Brainprint{
		neuronMap: make(map[string]*Neuron),
		linkMap:   make(map[string]*Link),
	}
}

func (b *Brainprint) DeepCopy() *Brainprint {
	if b == nil {
		return nil
	}
	cp := &Brainprint{
		neuronMap: make(map[string]*Neuron),
		linkMap:   make(map[string]*Link),
	}
	if b.timeout != nil {
		timeout := *b.timeout
		cp.timeout = &timeout
	}
	for id, neuron := range b.neuronMap {
		cp.neuronMap[id] = neuron.DeepCopy()
	}
	for id, link := range b.linkMap {
		cp.linkMap[id] = link.DeepCopy()
	}
	return cp
}

func (b *Brainprint) AddNeuron(processFn func(Brain) error) string {
	neuron := newNeuron()
	neuron.bindProcessor(&DefaultProcessor{processFn: processFn})
	b.neuronMap[neuron.id] = neuron
	return neuron.id
}

func (b *Brainprint) AddNeuronWithProcessor(p Processor) string {
	neuron := newNeuron()
	neuron.bindProcessor(p)
	b.neuronMap[neuron.id] = neuron
	return neuron.id
}

func (b *Brainprint) GetNeuron(id string) *Neuron {
	return b.neuronMap[id]
}

func (b *Brainprint) AddLink(fromID, toID string) (string, error) {
	// check neuron exist in brain
	if fromID == EndNeuronID {
		return "", fmt.Errorf("END neuron cannot conduct to any neuron")
	}
	if toID == EndNeuronID {
		b.ensureEndNeuron()
	}
	from := b.GetNeuron(fromID)
	if from == nil {
		return "", errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", fromID)
	}
	to := b.GetNeuron(toID)
	if to == nil {
		return "", errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", toID)
	}

	link := newLink(from, to)

	if err := to.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}
	if err := from.addLinkToDefaultConductGroup(link); err != nil {
		// rollback
		to.deleteTriggerGroup(link)
		// error
		return "", errors.Wrapf(err, "add link to default conduct group error")
	}
	b.linkMap[link.id] = link

	return link.id, nil
}

func (b *Brainprint) GetLink(id string) *Link {
	return b.linkMap[id]
}

func (b *Brainprint) AddEntryLink(toID string) (string, error) {
	to := b.GetNeuron(toID)
	if to == nil {
		return "", errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", toID)
	}

	link := newEntryLink(to)

	if err := to.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}

	b.linkMap[link.id] = link

	return link.id, nil
}

// AddEndLink add link from specific neuron to END neuron, if END neuron not exist, create it.
func (b *Brainprint) AddEndLink(fromID string) (string, error) {
	b.ensureEndNeuron()
	end := b.GetNeuron(EndNeuronID)
	from := b.GetNeuron(fromID)
	if from == nil {
		return "", errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", fromID)
	}

	link := newLink(from, end)

	if err := end.addTriggerGroup(link); err != nil {
		// rollback, do nothing

		// error
		return "", errors.Wrapf(err, "add trigger group with link error")
	}
	if err := from.addLinkToDefaultConductGroup(link); err != nil {
		// rollback
		end.deleteTriggerGroup(link)
		// error
		return "", errors.Wrapf(err, "add link to default conduct group error")
	}

	b.linkMap[link.id] = link

	return link.id, nil
}

// AddLinkToConductGroup add links to specific named conduct group.
// if group not exist, create the group. Groups that allow empty links.
// The specified link will remove from the default group, if it originally belonged to the default group.
func (b *Brainprint) AddLinkToConductGroup(neuronID string, groupName string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.Wrapf(ErrorLinkNotFound, "link ID %s", id)
		}
		links = append(links, link)
	}

	return neu.addLinkToConductGroup(groupName, links...)
}

// DeleteConductGroup ...
func (b *Brainprint) DeleteConductGroup(neuronID string, groupName string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", neuronID)
	}

	return neu.deleteConductGroup(groupName)
}

// AddTriggerGroup ...
func (b *Brainprint) AddTriggerGroup(neuronID string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.Wrapf(ErrorLinkNotFound, "link ID %s", id)
		}
		links = append(links, link)
	}

	return neu.addTriggerGroup(links...)
}

// DeleteTriggerGroup ...
func (b *Brainprint) DeleteTriggerGroup(neuronID string, linkIDs ...string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", neuronID)
	}

	links := make([]*Link, 0)
	for _, id := range linkIDs {
		link := b.GetLink(id)
		if link == nil {
			return errors.Wrapf(ErrorLinkNotFound, "link ID %s", id)
		}
		links = append(links, link)
	}

	neu.deleteTriggerGroup(links...)

	return nil
}

// Build will build BrainLocal
func (b *Brainprint) Build(withOpts ...Option) Brain {
	bpcp := b.DeepCopy()
	fmt.Printf("brainprint copy: %s\n", bpcp)
	brain := NewBrainLocal(*bpcp, withOpts...)

	return brain
}

// BindConductGroupSelectFunc bind custom select function of conduct group, default select default conduct group.
func (b *Brainprint) BindConductGroupSelectFunc(neuronID string, selectFn func(brain Brain) string) error {
	neu := b.GetNeuron(neuronID)
	if neu == nil {
		return errors.Wrapf(ErrorNeuronNotFound, "neuron ID %s", neuronID)
	}

	neu.selectFn = selectFn

	return nil
}

func (b *Brainprint) HasEndNeuron() bool {
	for nid, _ := range b.neuronMap {
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
	b.neuronMap[neuron.id] = neuron
}

func (b *Brainprint) String() string {
	neuronMapString := utils.PrintMap(b.neuronMap)
	linkMapString := utils.PrintMap(b.linkMap)

	return fmt.Sprintf(`{
		"neuron_map": %s,
		"link_map": %s,
		"timeout": "%s"
	}`, neuronMapString, linkMapString, b.timeout)

}
