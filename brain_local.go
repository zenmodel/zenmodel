package zenmodel

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func NewBrainLocal(brainprint Brainprint, withOpts ...Option) *BrainLocal {
	brain := &BrainLocal{
		Brainprint: brainprint,
		brainLocalOptions: brainLocalOptions{
			rateLimiterBaseDelay: 5 * time.Millisecond,
			rateLimiterMaxDelay:  1000 * time.Second,
			workerNum:            5,
		},
		queue: nil,
		state: BrainStateSleeping,
	}

	// maintainer
	for _, opt := range withOpts {
		opt.apply(brain)
	}

	return brain
}

type BrainLocal struct {
	Brainprint
	// read only state, unable to set
	// Brainprint is in the Running state when there are 1 or more Activate neuron
	// or 1 or more StandBy Edge.
	state BrainState
	// brain context
	context Context

	brainLocalOptions
	// local maintainer queue
	queue workqueue.RateLimitingInterface
}

type brainLocalOptions struct {
	rateLimiterBaseDelay time.Duration
	rateLimiterMaxDelay  time.Duration
	workerNum            int
}

func (b *BrainLocal) Entry() {
	// get all entry links
	linkIDs := make([]string, 0)
	for _, link := range b.linkMap {
		if link.IsEntryLink() {
			linkIDs = append(linkIDs, link.id)
		}
	}

	b.trigLinks(linkIDs...)

	return
}

func (b *BrainLocal) TrigLinks(linkIDs ...string) {
	b.trigLinks(linkIDs...)

	return
}

func (b *BrainLocal) Start() {
	// re-new queue
	if b.queue != nil {
		b.queue.ShutDown()
	}
	b.queue = workqueue.NewRateLimitingQueueWithConfig(
		workqueue.NewItemExponentialFailureRateLimiter(b.rateLimiterBaseDelay, b.rateLimiterMaxDelay),
		workqueue.RateLimitingQueueConfig{
			Name: "brain-queue-local", // TODO add brain ID ?
		})

	// start worker, will terminate goroutine when queue shutdown
	for i := 0; i < b.workerNum; i++ {
		go b.runWorker()
	}
}

func (b *BrainLocal) Shutdown() {
	b.queue.ShutDown()
}

func (b *BrainLocal) SendMessage(message Message) {
	b.queue.Add(message)
}

func (b *BrainLocal) runWorker() {
	//m.logger.Info("run worker")
	for b.processFromQueue() {
	}
}

// should block, without goroutine
func (b *BrainLocal) processFromQueue() bool {
	msg, shutdown := b.queue.Get()
	if shutdown {
		return false
	}
	defer b.queue.Done(msg)

	message := msg.(Message)
	switch message.kind {
	case MessageKindLink:
		// 1. link message
		if err := b.HandleLink(message.Action, message.ID); err != nil {
			// TODO log error
			return true
		}

	case MessageKindNeuron:
		// 2. neuron message
		if err := b.HandleNeuron(message.Action, message.ID); err != nil {
			// TODO log error
			return true
		}

	default:
		// log unsupported message kind
		return true
	}

	// 3. 重新计算 brain 状态并刷新
	b.RefreshState()

	b.queue.Forget(msg)
	return true
}

func (b *BrainLocal) trigLinks(linkIDs ...string) {
	if len(linkIDs) == 0 {
		return
	}

	// ensure brain maintainer start
	b.ensureStart()

	// goroutine wait
	var wg sync.WaitGroup
	for _, linkID := range linkIDs {
		link := b.GetLink(linkID)
		if link == nil {
			continue
		}

		wg.Add(1)
		go b.trigLink(&wg, link)
	}
	wg.Wait()

	return
}

func (b *BrainLocal) ensureStart() {
	if b.state == BrainStateSleeping {
		b.Start()
		b.state = BrainStateAwake
	}

	return
}

func (b *BrainLocal) trigLink(wg *sync.WaitGroup, link *Link) {
	defer wg.Done()

	if link.state != LinkStateReady {
		// change link state as ready
		link.state = LinkStateReady

		// send link message
		b.SendMessage(Message{
			kind:   MessageKindLink,
			Action: MessageActionLinkReady,
			ID:     link.id,
		})
	}

	return
}

func (b *BrainLocal) RefreshState() {
	_, activateCnt := b.getNeuronCountByState()
	_, waitCnt, readyCnt := b.getLinkCountByState()
	// set to running

	// set to sleeping, and shutdown maintainer
	if activateCnt+waitCnt+readyCnt == 0 {
		b.Shutdown()
		b.state = BrainStateSleeping
	} else { // > 0
		b.state = BrainStateRunning
	}

}

func (b *BrainLocal) getNeuronCountByState() (int, int) {
	var inhibitCnt, activateCnt int
	for _, neuron := range b.neuronMap {
		switch neuron.state {
		case NeuronStateInhibited:
			inhibitCnt++
		case NeuronStateActivated:
			activateCnt++
		}
	}

	return inhibitCnt, activateCnt
}

func (b *BrainLocal) getLinkCountByState() (int, int, int) {
	var initCnt, waitCnt, readyCnt int
	for _, link := range b.linkMap {
		switch link.state {
		case LinkStateInit:
			initCnt++
		case LinkStateWait:
			waitCnt++
		case LinkStateReady:
			readyCnt++
		}
	}

	return initCnt, waitCnt, readyCnt
}

func (b *BrainLocal) HandleLink(action MessageAction, linkID string) error {
	link := b.GetLink(linkID)
	// TODO link nil error
	switch action {
	case MessageActionLinkInit:
		// do nothing
	case MessageActionLinkWait:
		// do nothing
	case MessageActionLinkReady:
		destNeuron := b.GetNeuron(link.to)
		// TODO  neuron nil error
		// try dest neuron activate
		b.SendMessage(Message{
			kind:   MessageKindNeuron,
			Action: MessageActionNeuronTryActivate,
			ID:     destNeuron.id,
		})

	default:
		// TODO unsupported link action error

	}

	return nil
}

func (b *BrainLocal) HandleNeuron(action MessageAction, neuronID string) error {
	neuron := b.GetNeuron(neuronID)
	// TODO neuron nil error

	switch action {
	case MessageActionNeuronTryInhibit:
		// do nothing , neuron 只有自己 Inhibit
	case MessageActionNeuronTryActivate:
		return b.tryActivateNeuron(neuron)
	default:
		// TODO unsupported neuron action error
	}

	return nil
}

func (b *BrainLocal) tryActivateNeuron(neuron *Neuron) error {
	if neuron.state == NeuronStateActivated {
		// TODO log already activated
		return nil
	}

	should := b.ifNeuronShouldActivate(neuron)
	if !should {
		// TODO log
		return nil
	}

	return b.activateNeuron(neuron)
}

// ifNeuronShouldActivate return link IDs in trigged trigger group
func (b *BrainLocal) ifNeuronShouldActivate(neuron *Neuron) bool {
	if b.state == BrainStateSleeping {
		return false
	}

	// 如果任一触发组中的 link 全都是 Ready, 则应该 activate neuron
	for _, group := range neuron.triggerGroups {
		trigLinks := make([]*Link, 0)
		for _, linkID := range group {
			link := b.GetLink(linkID)
			if link.state == LinkStateReady {
				trigLinks = append(trigLinks, link)
			} else {
				break
			}
		}
		if len(trigLinks) == len(group) && len(group) != 0 {
			return true
		}
	}

	return false
}

func (b *BrainLocal) activateNeuron(neuron *Neuron) error {
	neuron.state = NeuronStateActivated
	// in-link set init
	for _, group := range neuron.triggerGroups {
		for _, linkID := range group {
			link := b.GetLink(linkID)
			link.state = LinkStateInit
		}
	}

	// out-link set wait
	for _, group := range neuron.conductGroups {
		for linkID, _ := range group {
			link := b.GetLink(linkID)
			link.state = LinkStateWait
		}
	}

	neuron.count.process++
	err := neuron.processor.Process(b)
	neuron.state = NeuronStateInhibited
	if err != nil {
		neuron.count.failed++
		return err
	}

	// SucceedCount++
	neuron.count.succeed++
	// 决策出边/传导组
	selected := neuron.selectFn(b)
	// 选中的 conduct group 中的 link 状态为 wait 的设置为 ready，SendMessage （为 init 的则不改变）
	for linkID, _ := range neuron.conductGroups[selected] {
		link := b.GetLink(linkID)
		if link.state == LinkStateWait {
			link.state = LinkStateReady
			b.SendMessage(Message{
				kind:   MessageKindLink,
				Action: MessageActionLinkReady,
				ID:     link.id,
			})
		}
	}

	return nil
}
