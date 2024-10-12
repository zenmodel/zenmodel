package brainlocal

import (
	"github.com/zenmodel/zenmodel/core"
	"github.com/zenmodel/zenmodel/internal/utils"
	"github.com/zenmodel/zenmodel/processor"
)

type neuron struct {
	id     string
	labels map[string]string
	spec   neuronSpec
	status neuronStatus
}

type neuronSpec struct {
	processor processor.Processor
	// 触发组,触发组是用来控制 Neuron 的触发条件
	triggerGroups map[string][]*link
	// 传播组,传播组是用来控制 Neuron 之间的传播关系
	castGroups map[string][]*link
	// 在 neuron 运行成功之后通过 Selector 决定传导到哪一个传播组
	selector processor.Selector
}

type neuronStatus struct {
	state core.NeuronState
	count struct {
		process int
		succeed int
		failed  int
	}
}

func newNeuron(n core.Neuron, linkMap map[string]*link) *neuron {
	neu := &neuron{
		id:     n.GetID(),
		labels: utils.LabelsDeepCopy(n.GetLabels()),
		spec: neuronSpec{
			processor:     n.GetProcessor(),
			selector:      n.GetSelector(),
			triggerGroups: make(map[string][]*link),
			castGroups:    make(map[string][]*link),
		},
		status: neuronStatus{
			state: core.NeuronStateInactive,
		},
	}

	for gName, links := range n.ListTriggerGroups() {
		neu.spec.triggerGroups[gName] = make([]*link, len(links))
		for i, linkID := range links {
			neu.spec.triggerGroups[gName][i] = linkMap[linkID]
		}
	}

	for gName, links := range n.ListCastGroups() {
		neu.spec.castGroups[gName] = make([]*link, len(links))
		for i, linkID := range links {
			neu.spec.castGroups[gName][i] = linkMap[linkID]
		}
	}

	return neu
}
