package zenmodel

import (
	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/core"
	"github.com/zenmodel/zenmodel/internal/utils"
)

func newLink(srcNeuronID, destNeuronID string) *link {
	return &link{
		id:     utils.GenIDShort(),
		labels: make(map[string]string),
		src:    srcNeuronID,
		dest:   destNeuronID,
	}
}

func newEntryLink(destNeuronID string) *link {
	return &link{
		id:     utils.GenIDShort(),
		labels: make(map[string]string),
		src:    core.EntryLinkFrom,
		dest:   destNeuronID,
	}
}

func newEndLink(srcNeuronID string) *link {
	return &link{
		id:     utils.GenIDShort(),
		labels: make(map[string]string),
		src:    srcNeuronID,
		dest:   core.EndLinkTo,
	}
}

type link struct {
	// ID
	id string
	// labels
	labels map[string]string
	// from source neuron ID
	src string
	// to destination neuron ID
	dest string
}

func (l *link) GetSrcNeuronID() string {
	return l.src
}

func (l *link) GetDestNeuronID() string {
	return l.dest
}

func (l *link) GetID() string {
	return l.id
}

func (l *link) GetLabels() map[string]string {
	return l.labels
}

func (l *link) SetLabels(labels map[string]string) {
	l.labels = labels
}

func (l *link) IsEntryLink() bool {
	return l.src == core.EntryLinkFrom
}

func (l *link) IsEndLink() bool {
	return l.dest == core.EndLinkTo
}

func (l *link) deepCopy() *link {
	return &link{
		id:     l.id,
		labels: utils.LabelsDeepCopy(l.labels),
		src:    l.src,
		dest:   l.dest,
	}
}

func (l *link) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", l.id).
		Any("labels", l.labels).
		Str("src", l.src).
		Str("dest", l.dest)
}
