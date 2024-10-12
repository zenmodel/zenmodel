package brainlite

import (
	"github.com/rs/zerolog"
)

// Option configures a BrainLite in build.
type Option interface {
	apply(brain *BrainLite)
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(*BrainLite)

func (f optionFunc) apply(brain *BrainLite) {
	f(brain)
}

// WithNeuronWorkerNum sets the neuron process worker number
func WithNeuronWorkerNum(workerNun int) Option {
	return optionFunc(func(brain *BrainLite) {
		brain.nWorkerNum = workerNun
	})
}

// WithNeuronQueueLen sets the neuron process queue length
func WithNeuronQueueLen(nQueueLen int) Option {
	return optionFunc(func(brain *BrainLite) {
		brain.nQueueLen = nQueueLen
	})
}

// WithLoggerLevel sets the default logger with specific level
func WithLoggerLevel(level zerolog.Level) Option {
	return optionFunc(func(brain *BrainLite) {
		brain.logger = brain.logger.Level(level)
	})
}

// WithLogger sets the specific logger
func WithLogger(logger zerolog.Logger) Option {
	return optionFunc(func(brain *BrainLite) {
		brain.logger = logger
	})
}

// WithID sets the specific brain ID
func WithID(brainID string) Option {
	return optionFunc(func(brain *BrainLite) {
		brain.id = brainID
	})
}
