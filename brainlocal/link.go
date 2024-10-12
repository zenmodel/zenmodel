package brainlocal

import (
	"github.com/zenmodel/zenmodel/core"
)

type link struct {
	id     string
	spec   linkSpec
	status linkStatus
}

type linkSpec struct {
	// from neuron ID
	from string
	// to neuron ID
	to string
}

type linkStatus struct {
	state core.LinkState
	count struct {
		// from 执行完整，开始尝试传递的次数
		process int
		// 传递成功的次数
		succeed int
		// 传递失败的次数。可能是超时、其他触发组触发当前的取消等原因
		failed int
	}
}

func newLink(l core.Link) *link {
	return &link{
		id: l.GetID(),
		spec: linkSpec{
			from: l.GetSrcNeuronID(),
			to:   l.GetDestNeuronID(),
		},
		status: linkStatus{
			state: core.LinkStateInit,
		},
	}
}

func (l *link) isEntryLink() bool {
	if l.spec.from == core.EntryLinkFrom {
		return true
	}

	return false
}
