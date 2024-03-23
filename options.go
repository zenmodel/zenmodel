package zenmodel

import (
	"time"

	"github.com/zenmodel/zenmodel/internal/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Option configures a BrainLocal in build.
type Option interface {
	apply(brain Brain)
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(Brain)

func (f optionFunc) apply(brain Brain) {
	f(brain)
}

// WithLocalMaintainer sets the rate limiter options for the BrainLocal maintainer.
func WithLocalMaintainer(rateLimiterBaseDelay, rateLimiterMaxDelay time.Duration) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.brainLocalOptions = brainLocalOptions{
			rateLimiterBaseDelay: rateLimiterBaseDelay,
			rateLimiterMaxDelay:  rateLimiterMaxDelay,
		}
	})
}

// WithWorkerNum sets the worker number for the BrainLocal implementation.
func WithWorkerNum(workerNun int) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.brainLocalOptions = brainLocalOptions{
			workerNum: workerNun,
		}
	})
}

// WithMemorySetting sets the memory setting for BrainLocal implementation.
func WithMemorySetting(memoryNumCounters, memoryMaxCost int64) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.brainLocalOptions = brainLocalOptions{
			memoryNumCounters: memoryNumCounters,
			memoryMaxCost:     memoryMaxCost,
		}
	})
}

// WithLoggerLevel sets the default logger with specific level for BrainLocal implementation.
func WithLoggerLevel(level zapcore.Level) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.logger = log.NewDefaultLoggerWithLevel(level)
	})
}

// WithLogger sets the specific logger for BrainLocal implementation.
func WithLogger(logger *zap.Logger) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.logger = logger
	})
}

// WithID sets the specific brain ID for BrainLocal implementation.
func WithID(brainID string) Option {
	return optionFunc(func(brain Brain) {
		brainLocal := brain.(*BrainLocal)
		brainLocal.id = brainID
	})
}

// NeuronOption configures a neuron.
type NeuronOption interface {
	apply(neuron *Neuron)
}

// neuronOptionFunc wraps a func, so it satisfies the NeuronOption interface.
type neuronOptionFunc func(*Neuron)

func (f neuronOptionFunc) apply(neuron *Neuron) {
	f(neuron)
}

// WithLabels sets the specific labels for Neuron
func WithLabels(labels map[string]string) NeuronOption {
	return neuronOptionFunc(func(neuron *Neuron) {
		neuron.labels = labels
	})
}

// WithSelectFn sets the specific selectFn for Neuron
func WithSelectFn(selectFn func(brain Brain) string) NeuronOption {
	return neuronOptionFunc(func(neuron *Neuron) {
		neuron.selectFn = selectFn
	})
}
