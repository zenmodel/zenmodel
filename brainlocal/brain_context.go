package brainlocal

type brainContext struct {
	b               *BrainLocal
	currentNeuronID string
}

func (c *brainContext) SetMemory(keysAndValues ...interface{}) error {
	return c.b.SetMemory(keysAndValues...)
}

func (c *brainContext) GetMemory(key interface{}) interface{} {
	return c.b.GetMemory(key)
}

func (c *brainContext) ExistMemory(key interface{}) bool {
	return c.b.ExistMemory(key)
}

func (c *brainContext) DeleteMemory(key interface{}) {
	c.b.DeleteMemory(key)
}

func (c *brainContext) ClearMemory() {
	c.b.ClearMemory()
}

func (c *brainContext) GetCurrentNeuronID() string {
	return c.currentNeuronID
}

func (c *brainContext) ContinueCast() {
	_, ok := c.b.neurons[c.currentNeuronID]
	if !ok {
		return
	}

	c.b.publishEvent(maintainEvent{
		kind:   eventKindNeuron,
		action: eventActionNeuronCastAnyway,
		id:     c.currentNeuronID,
	})
}
