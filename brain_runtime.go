package zenmodel

import (
	"github.com/zenmodel/zenmodel/internal/constants"
)

type BrainRuntime interface {
	SetMemory(keysAndValues ...interface{}) error
	GetMemory(key interface{}) (interface{}, bool)
	DeleteMemory(key interface{})
	ContinueCast()
}

type brainLocalRuntime struct {
	brain         *BrainLocal
	currentNeuron *Neuron
}

func (r *brainLocalRuntime) SetMemory(keysAndValues ...interface{}) error {
	return r.brain.SetMemory(keysAndValues...)
}

func (r *brainLocalRuntime) GetMemory(key interface{}) (interface{}, bool) {
	return r.brain.GetMemory(key)
}

func (r *brainLocalRuntime) DeleteMemory(key interface{}) {
	r.brain.DeleteMemory(key)
}

func (r *brainLocalRuntime) ContinueCast() {
	if r.currentNeuron == nil {
		return
	}
	r.brain.SendMessage(constants.Message{
		Kind:   constants.MessageKindNeuron,
		Action: constants.MessageActionNeuronCastAnyway,
		ID:     r.currentNeuron.id,
	})
}
