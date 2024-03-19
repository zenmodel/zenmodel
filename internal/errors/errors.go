package zenmodel

import (
	"github.com/pkg/errors"
)

var (
	ErrorNeuronNotFound = errors.New("neuron not found")
	ErrorLinkNotFound   = errors.New("link not found")
)
