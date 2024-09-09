package brainlocal

import (
	"github.com/rs/zerolog"
)

// Option configures a BrainLocal in build.
type Option interface {
	apply(brain *BrainLocal)
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(*BrainLocal)

func (f optionFunc) apply(brain *BrainLocal) {
	f(brain)
}

// WithNeuronWorkerNum sets the neuron process worker number
func WithNeuronWorkerNum(workerNun int) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.nWorkerNum = workerNun
	})
}

// WithNeuronQueueLen sets the neuron process queue length
func WithNeuronQueueLen(nQueueLen int) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.nQueueLen = nQueueLen
	})
}

// WithMemorySetting sets the memory setting
func WithMemorySetting(memoryNumCounters, memoryMaxCost int64) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.numCounters = memoryNumCounters
		brain.maxCost = memoryMaxCost
	})
}

// WithLoggerLevel sets the default logger with specific level
func WithLoggerLevel(level zerolog.Level) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.logger = brain.logger.Level(level)
	})
}

// WithLogger sets the specific logger
func WithLogger(logger zerolog.Logger) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.logger = logger
	})
}

// WithID sets the specific brain ID
func WithID(brainID string) Option {
	return optionFunc(func(brain *BrainLocal) {
		brain.id = brainID
	})
}
