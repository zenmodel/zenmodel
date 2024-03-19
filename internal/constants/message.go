package constants

import "go.uber.org/zap/zapcore"

type Message struct {
	Kind   MessageKind
	Action MessageAction
	ID     string
}

func (m Message) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("kind", string(m.Kind))
	enc.AddString("action", string(m.Action))
	enc.AddString("id", m.ID)

	return nil
}

type MessageKind string
type MessageAction string

const (
	MessageKindNeuron MessageKind = "neuron"
	MessageKindLink   MessageKind = "link"
	MessageKindBrain  MessageKind = "brain"

	MessageActionLinkInit          MessageAction = "link_init"
	MessageActionLinkReady         MessageAction = "link_ready"
	MessageActionLinkWait          MessageAction = "link_wait"
	MessageActionNeuronTryActivate MessageAction = "try_activate_neuron"
	MessageActionNeuronTryInhibit  MessageAction = "try_inhibit_neuron"
	MessageActionBrainSleep        MessageAction = "brain_sleep"
)
