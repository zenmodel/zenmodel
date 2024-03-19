package zenmodel

import (
	"github.com/zenmodel/zenmodel/internal/utils"
	"go.uber.org/zap/zapcore"
)

const (
	EntryLinkFrom = "EXTERNAL_SIGNAL"
	EndLinkTo     = EndNeuronID
)

type Link struct {
	// ID 不可编辑
	id string
	// state 不可编辑，
	state LinkState
	count struct {
		// from 执行完整，开始尝试传递的次数
		process int
		// 传递成功的次数
		succeed int
		// 传递失败的次数。可能是超时、其他触发组触发当前的取消等原因
		failed int
	}

	// from neuron ID
	from string
	// to neuron ID
	to string
}

type LinkState string

const (
	LinkStateInit  LinkState = "Init"
	LinkStateWait  LinkState = "Wait"
	LinkStateReady LinkState = "Ready"
)

func newLink(from, to *Neuron) *Link {
	return &Link{
		id:    utils.GenUUID(),
		state: LinkStateInit,
		from:  from.id,
		to:    to.id,
	}
}

func newEntryLink(to *Neuron) *Link {
	return &Link{
		id:    utils.GenUUID(),
		state: LinkStateInit,
		from:  EntryLinkFrom,
		to:    to.id,
	}
}

func (l *Link) DeepCopy() *Link {
	if l == nil {
		return nil
	}
	cp := &Link{
		id:    l.id,
		state: l.state,
		from:  l.from,
		to:    l.to,
	}
	cp.count.process = l.count.process
	cp.count.succeed = l.count.succeed
	cp.count.failed = l.count.failed

	return cp
}

func (l *Link) IsEntryLink() bool {
	if l.from == EntryLinkFrom {
		return true
	}

	return false
}

func (l *Link) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("id", l.id)
	enc.AddString("state", string(l.state))
	enc.AddString("src", l.from)
	enc.AddString("dest", l.to)

	return nil
}
