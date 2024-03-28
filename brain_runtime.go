package zenmodel

import (
	"github.com/zenmodel/zenmodel/internal/constants"
)

type BrainRuntime interface {
	// SetMemory set memories for brain, one key value pair is one memory.
	// memory will lazy initial util `SetMemory` or any link trig
	SetMemory(keysAndValues ...interface{}) error
	// GetMemory get memory by key
	GetMemory(key interface{}) interface{}
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key interface{}) bool
	// DeleteMemory delete one memory by key
	DeleteMemory(key interface{})
	// ClearMemory clear all memories
	ClearMemory()
	ContinueCast()
}

type brainLocalRuntime struct {
	brain         *BrainLocal
	currentNeuron *Neuron
}

func (r *brainLocalRuntime) SetMemory(keysAndValues ...interface{}) error {
	return r.brain.SetMemory(keysAndValues...)
}

func (r *brainLocalRuntime) GetMemory(key interface{}) interface{} {
	return r.brain.GetMemory(key)
}

func (r *brainLocalRuntime) ExistMemory(key interface{}) bool {
	return r.brain.ExistMemory(key)
}

func (r *brainLocalRuntime) DeleteMemory(key interface{}) {
	r.brain.DeleteMemory(key)
}

func (r *brainLocalRuntime) ClearMemory() {
	r.brain.ClearMemory()
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
