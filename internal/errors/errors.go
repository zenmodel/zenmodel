package errors

import (
	"github.com/pkg/errors"
	"github.com/zenmodel/zenmodel/internal/constants"
)

var (
	errNeuronNotFound           = errors.New("neuron not found")
	errLinkNotFound             = errors.New("link not found")
	errUnsupportedMessageKind   = errors.New("unsupported message kind")
	errUnsupportedMessageAction = errors.New("unsupported message action")
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

func ErrUnsupportedMessageKind(messageKind constants.MessageKind) error {
	return errors.Wrapf(errUnsupportedMessageKind, "kind: %s", messageKind)
}

func ErrUnsupportedMessageAction(messageAction constants.MessageAction) error {
	return errors.Wrapf(errUnsupportedMessageAction, "action: %s", messageAction)
}
