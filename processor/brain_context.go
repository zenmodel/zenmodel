package processor

type BrainContext interface {
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
	// GetCurrentNeuronID get current neuron id
	GetCurrentNeuronID() string
	// GetCurrentNeuronLabels get current neuron labels
	GetCurrentNeuronLabels() map[string]string
	// GetBrainID get brain id
	GetBrainID() string 
	// GetBrainLabels get brain labels
	GetBrainLabels() map[string]string
	// ContinueCast keep current process running, and continue cast
	ContinueCast()
	// TODO Context 继承 context.Context
	//context.Context
}

type BrainContextReader interface {
	// GetMemory get memory by key
	GetMemory(key interface{}) interface{}
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key interface{}) bool
	// GetCurrentNeuronID get current neuron id
	GetCurrentNeuronID() string
	// TODO Context 继承 context.Context
	//context.Context
}
