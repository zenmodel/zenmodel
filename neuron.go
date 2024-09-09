package zenmodel

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/brain"
	"github.com/zenmodel/zenmodel/internal/errors"
	"github.com/zenmodel/zenmodel/internal/utils"
	"github.com/zenmodel/zenmodel/processor"
)

func newNeuron(p processor.Processor) *neuron {
	n := &neuron{
		id:            utils.GenIDShort(),
		labels:        make(map[string]string),
		processor:     p,
		triggerGroups: make(triggerGroups),
		castGroups:    make(castGroups),
		selector:      &processor.DefaultSelector{},
	}

	return n
}

func newEndNeuron() *neuron {
	n := &neuron{
		id:            brain.EndNeuronID,
		labels:        make(map[string]string),
		processor:     &processor.EmptyProcessor{},
		triggerGroups: make(triggerGroups),
	}

	return n
}

type neuron struct {
	// ID
	id string
	// labels
	labels map[string]string
	// processor 处理器
	processor processor.Processor
	// 触发组,触发组是用来控制 Neuron 的触发条件
	// key: group ID, value: list of link ID
	triggerGroups triggerGroups
	// 传播组,传播组是用来控制 Neuron 之间的传播关系
	// key: group ID/Name, value: map of link ID
	castGroups castGroups
	// 在 neuron 运行成功之后通过 Selector 决定传导到哪一个传播组
	selector processor.Selector
}

func (n *neuron) deepCopy() *neuron {
	return &neuron{
		id:            n.id,
		labels:        utils.LabelsDeepCopy(n.labels),
		processor:     n.processor,
		triggerGroups: n.triggerGroups.deepCopy(),
		castGroups:    n.castGroups.deepCopy(),
		selector:      n.selector,
	}
}

func (n *neuron) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", n.id).
		Interface("labels", n.labels).
		Interface("triggerGroups", n.triggerGroups).
		Interface("castGroups", n.castGroups.format())
}

type castGroups map[string]map[string]struct{}

func (cgs castGroups) deepCopy() castGroups {
	newMap := make(castGroups)

	for key, value := range cgs {
		newInnerMap := make(map[string]struct{})
		for innerKey, innerValue := range value {
			newInnerMap[innerKey] = innerValue
		}
		newMap[key] = newInnerMap
	}

	return newMap
}

func (cgs castGroups) format() map[string][]string {
	newMap := make(map[string][]string)

	for key, value := range cgs {
		var newSlice []string
		for innerKey := range value {
			newSlice = append(newSlice, innerKey)
		}

		newMap[key] = newSlice
	}

	return newMap
}

type triggerGroups map[string][]string

func (tgs triggerGroups) deepCopy() triggerGroups {
	newGs := make(triggerGroups)

	for key, value := range tgs {
		newSlice := make([]string, len(value))
		copy(newSlice, value)
		newGs[key] = newSlice
	}

	return newGs
}

func (n *neuron) GetID() string {
	return n.id
}

func (n *neuron) GetLabels() map[string]string {
	return n.labels
}

func (n *neuron) GetProcessor() processor.Processor {
	return n.processor
}

func (n *neuron) GetSelector() processor.Selector {
	return n.selector
}

func (n *neuron) ListInLinkIDs() []string {
	linkMap := make(map[string]struct{})
	for _, group := range n.triggerGroups {
		for _, l := range group {
			linkMap[l] = struct{}{}
		}
	}
	links := make([]string, 0, len(linkMap))
	for l, _ := range linkMap {
		links = append(links, l)
	}

	return links
}

func (n *neuron) ListOutLinkIDs() []string {
	linkMap := make(map[string]struct{})
	for _, group := range n.castGroups {
		for l, _ := range group {
			linkMap[l] = struct{}{}
		}
	}
	links := make([]string, 0, len(linkMap))
	for l, _ := range linkMap {
		links = append(links, l)
	}

	return links
}

func (n *neuron) ListTriggerGroups() map[string][]string {
	return n.triggerGroups.deepCopy()
}

func (n *neuron) ListCastGroups() map[string][]string {
	return n.castGroups.format()
}

func (n *neuron) SetLabels(labels map[string]string) {
	n.labels = labels
}

// AddTriggerGroup in-link 连入 neuron 之后, 默认自成一组, 即一条 in-link 划分在一个 trigger group 中,
// 也就是说默认情况下任意一条 in-link 都可以触发 neuron.
// AddTriggerGroup 用来将指定 links 划入同一个 trigger group 中,
// 如果新划分的 trigger group 包含了存量的 trigger group ，那存量的 trigger group 将被移除，
// 如果新划分的 trigger group 被存量的 trigger group 包含，那么不会创建新划分的组，
// 因为只需要定义最大的触发条件，就会包含小的触发条件. 举例来说: 当 {A,B,C} 满足时 {A,B} 必定满足.
func (n *neuron) AddTriggerGroup(links ...brain.Link) error {
	if len(links) == 0 {
		return nil
	}
	for _, l := range links {
		if !n.hasInLink(l.GetID()) {
			return errors.ErrInLinkNotFound(l.GetID(), n.GetID())
		}
	}

	newGroup := make([]string, 0)
	for _, l := range links {
		newGroup = append(newGroup, l.GetID())
	}

	for key, group := range n.triggerGroups {
		// 新划分的 trigger group 被存量的 trigger group 包含，那么不会创建新划分的组
		if utils.SlicesContains(group, newGroup) {
			return nil
		}
		// 新划分的 trigger group 包含了存量的 trigger group ，那存量的 trigger group 将被移除
		if utils.SlicesContains(newGroup, group) {
			delete(n.triggerGroups, key)
		}
	}
	// add new group
	n.triggerGroups[utils.GenIDShort()] = newGroup

	return nil
}

// AddCastGroup out-link 连出 neuron 之后, 最初默认都在 default group, 即所有 out-link 划分在一个 default cast group 中,
// 也就是说默认所有 out-link 都会传播
// AddCastGroup 用来将指定 links 划入同一个 cast group 中,
// 指定 link 如果原本属于 default group，则先从 default group 中移除
// 指定 link 如果原本属于 其他非 default group，不会从其他 group 中移除
// 如果 groupName 已存在，则追加指定 link 划入该 group 中，该 group 原有的 link 不会变
func (n *neuron) AddCastGroup(groupName string, links ...brain.Link) error {
	if groupName == "" {
		return fmt.Errorf("group name is empty")
	}
	for _, l := range links {
		if !n.hasOutLink(l.GetID()) {
			return errors.ErrOutLinkNotFound(l.GetID(), n.GetID())
		}
	}
	// init
	if n.castGroups == nil {
		n.castGroups = map[string]map[string]struct{}{
			processor.DefaultCastGroupName: map[string]struct{}{},
		}
	}
	if n.castGroups[processor.DefaultCastGroupName] == nil {
		n.castGroups[processor.DefaultCastGroupName] = map[string]struct{}{}
	}
	if n.castGroups[groupName] == nil {
		n.castGroups[groupName] = map[string]struct{}{}
	}

	for _, l := range links {
		// 指定 link 如果原本属于 default group，则先从 default group 中移除
		_, ok := n.castGroups[processor.DefaultCastGroupName][l.GetID()]
		if ok {
			delete(n.castGroups[processor.DefaultCastGroupName], l.GetID())
		}
		// add link to group
		n.castGroups[groupName][l.GetID()] = struct{}{}
	}

	return nil
}

func (n *neuron) BindCastGroupSelectFunc(selectFn func(bcr processor.BrainContextReader) string) {
	n.bindCastGroupSelector(processor.NewFuncSelector(selectFn))
}

func (n *neuron) BindCastGroupSelector(selector processor.Selector) {
	n.bindCastGroupSelector(selector)
}

func (n *neuron) bindCastGroupSelector(selector processor.Selector) {
	n.selector = selector
}

func (n *neuron) addInLink(linkID string) {
	n.triggerGroups[utils.GenIDShort()] = []string{linkID}
}

// addOutLink 在 DEFAULT cast group 中添加 out-link
func (n *neuron) addOutLink(linkID string) {
	if _, ok := n.castGroups[processor.DefaultCastGroupName]; !ok {
		n.castGroups[processor.DefaultCastGroupName] = make(map[string]struct{})
	}
	n.castGroups[processor.DefaultCastGroupName][linkID] = struct{}{}
}

func (n *neuron) hasInLink(linkID string) bool {
	for _, group := range n.triggerGroups {
		for _, l := range group {
			if l == linkID {
				return true
			}
		}
	}

	return false
}

func (n *neuron) hasOutLink(linkID string) bool {
	for _, group := range n.castGroups {
		for l, _ := range group {
			if l == linkID {
				return true
			}
		}
	}

	return false
}
