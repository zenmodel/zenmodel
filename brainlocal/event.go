package brainlocal

import (
	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/core"
)

type maintainEvent struct {
	kind   eventKind
	action eventAction
	id     string
}

type eventKind string
type eventAction string

const (
	eventKindNeuron eventKind = "neuron"
	eventKindLink   eventKind = "link"
	eventKindBrain  eventKind = "brain"

	eventActionLinkInit          eventAction = "link_init"
	eventActionLinkReady         eventAction = "link_ready"
	eventActionLinkWait          eventAction = "link_wait"
	eventActionNeuronTryActivate eventAction = "try_activate_neuron"
	eventActionNeuronTryInactive eventAction = "try_inactive_neuron"
	eventActionNeuronTryCast     eventAction = "try_cast"
	eventActionNeuronCastAnyway  eventAction = "cast_anyway"
	eventActionBrainSleep        eventAction = "brain_sleep"
	eventActionBrainShutdown     eventAction = "brain_shutdown"
)

func (m maintainEvent) MarshalZerologObject(e *zerolog.Event) {
	e.Str("kind", string(m.kind)).
		Str("action", string(m.action)).
		Str("id", m.id)
}

func (b *BrainLocal) publishEvent(event maintainEvent) {
	if b.getState() == core.BrainStateShutdown || b.bQueue == nil { // 关闭中或没启动
		return
	}
	b.logger.Debug().Interface("event", event).Msg("publish maintain event")

	b.bQueue <- event
}
