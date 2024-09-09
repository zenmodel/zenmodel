package errors

import (
	"github.com/pkg/errors"
)

var (
	errNeuronNotFound = errors.New("neuron not found")
	errLinkNotFound   = errors.New("link not found")
)

func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

func ErrNeuronNotFound(neuronID string) error {
	return errors.Wrapf(errNeuronNotFound, "neuron: %s", neuronID)
}

func ErrLinkNotFound(linkID string) error {
	return errors.Wrapf(errLinkNotFound, "link: %s", linkID)
}

func ErrInLinkNotFound(linkID, neuronID string) error {
	return errors.Wrapf(errLinkNotFound, "in-link %s of neuron %s", linkID, neuronID)
}

func ErrOutLinkNotFound(linkID, neuronID string) error {
	return errors.Wrapf(errLinkNotFound, "out-link %s of neuron %s", linkID, neuronID)
}
