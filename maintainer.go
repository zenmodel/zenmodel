package zenmodel

type Maintainer interface {
	Start()
	Shutdown()
	SendMessage(Message)
}
type Message struct {
	kind   MessageKind
	Action MessageAction
	ID     string
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
