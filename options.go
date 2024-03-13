package zenmodel

import "time"

// Option configures a BrainLocal in build.
type Option interface {
	apply(brain Brain)
}

// optionFunc wraps a func, so it satisfies the Option interface.
type optionFunc func(Brain)

func (f optionFunc) apply(brain Brain) {
	f(brain)
}

func WithLocalMaintainer(rateLimiterBaseDelay, rateLimiterMaxDelay time.Duration) Option {
	return optionFunc(func(brain Brain) {
		brainlocal := brain.(*BrainLocal)
		brainlocal.brainLocalOptions = brainLocalOptions{
			rateLimiterBaseDelay: rateLimiterBaseDelay,
			rateLimiterMaxDelay:  rateLimiterMaxDelay,
		}
	})
}
