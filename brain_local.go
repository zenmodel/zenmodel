package zenmodel

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/zenmodel/zenmodel/internal/constants"
	"github.com/zenmodel/zenmodel/internal/errors"
	"github.com/zenmodel/zenmodel/internal/log"
	"github.com/zenmodel/zenmodel/internal/utils"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"
)

func NewBrainLocal(brainprint Brainprint, withOpts ...Option) *BrainLocal {
	brain := &BrainLocal{
		id:         utils.GenUUID(),
		Brainprint: brainprint,
		brainLocalOptions: brainLocalOptions{
			rateLimiterBaseDelay: 5 * time.Millisecond,
			rateLimiterMaxDelay:  1000 * time.Second,
			workerNum:            5,
			memoryNumCounters:    1e7,     // number of keys to track frequency of (10M).
			memoryMaxCost:        1 << 30, // maximum cost of cache (1GB).
		},
		memory: nil,
		queue:  nil,
		state:  BrainStateSleeping,
		logger: log.GetLogger(),
	}

	// reset maintainer, logger, ID, etc.
	for _, opt := range withOpts {
		opt.apply(brain)
	}

	brain.logger = brain.logger.With(zap.String("brain", brain.id))
	brain.logger.Debug("new local brain created", zap.Object("brainprint", &brainprint))

	return brain
}

type BrainLocal struct {
	// ID 不可编辑
	id string

	Brainprint
	// read only state, unable to set
	// Brainprint is in the Running state when there are 1 or more Activate neuron
	// or 1 or more StandBy Edge.
	state BrainState
	// brain memory
	memory *ristretto.Cache

	brainLocalOptions
	// local maintainer queue
	queue workqueue.RateLimitingInterface
	wg    sync.WaitGroup

	logger *zap.Logger
}

type brainLocalOptions struct {
	rateLimiterBaseDelay time.Duration
	rateLimiterMaxDelay  time.Duration
	workerNum            int
	memoryNumCounters    int64
	memoryMaxCost        int64
}

func (b *BrainLocal) Entry() error {
	// get all entry links
	linkIDs := make([]string, 0)
	for _, link := range b.links {
		if link.IsEntryLink() {
			linkIDs = append(linkIDs, link.id)
		}
	}

	return b.trigLinks(linkIDs...)
}

func (b *BrainLocal) EntryWithMemory(keysAndValues ...interface{}) error {
	if err := b.SetMemory(keysAndValues...); err != nil {
		return err
	}

	return b.Entry()
}
func (b *BrainLocal) TrigLinks(linkIDs ...string) error {
	return b.trigLinks(linkIDs...)
}

func (b *BrainLocal) SetMemory(keysAndValues ...interface{}) error {
	if len(keysAndValues)%2 != 0 {
		return fmt.Errorf("key and value are not paired")
	}
	if err := b.ensureMemoryInit(); err != nil {
		// TODO wrap error
		return err
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		b.memory.Set(k, v, 1) // TODO maybe calculate cost
		b.logger.Debug("set memory", zap.Any("key", k), zap.Any("value", v))
	}
	b.memory.Wait()

	return nil
}

func (b *BrainLocal) GetMemory(key interface{}) (interface{}, bool) {
	if b.memory == nil {
		return nil, false
	}

	return b.memory.Get(key)
}

func (b *BrainLocal) DeleteMemory(key interface{}) {
	if b.memory == nil {
		return
	}

	b.memory.Del(key)
}

func (b *BrainLocal) GetState() BrainState {
	return b.state
}

func (b *BrainLocal) Wait() {
	b.wg.Wait()
}

func (b *BrainLocal) Start() {
	b.logger.Info("brain maintainer start", zap.Int("workerNum", b.workerNum))

	// re-new queue
	if b.queue != nil {
		b.ShutDown()
	}
	b.queue = workqueue.NewRateLimitingQueueWithConfig(
		workqueue.NewItemExponentialFailureRateLimiter(b.rateLimiterBaseDelay, b.rateLimiterMaxDelay),
		workqueue.RateLimitingQueueConfig{
			Name: "brain-queue-local", // TODO add brain ID ?
		})

	// start worker, will terminate goroutine when queue shutdown
	for i := 0; i < b.workerNum; i++ {
		b.wg.Add(1)
		go b.runWorker()
	}
}

func (b *BrainLocal) ShutDown() {
	b.logger.Info("brain maintainer shutdown")
	b.queue.ShutDown()
}

func (b *BrainLocal) SendMessage(message constants.Message) {
	b.logger.Debug("send message", zap.Object("message", message))
	b.queue.Add(message)
}

func (b *BrainLocal) runWorker() {
	defer b.wg.Done()

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

	message := msg.(constants.Message)
	logger := b.logger.With(zap.Object("message", message))
	logger.Debug("process message")

	switch message.Kind {
	case constants.MessageKindLink:
		// 1. link message
		if err := b.HandleLink(message.Action, message.ID); err != nil {
			logger.Error("handle link message error", zap.Error(err))
			return true
		}
	case constants.MessageKindNeuron:
		// 2. neuron message
		if err := b.HandleNeuron(message.Action, message.ID); err != nil {
			logger.Error("handle neuron message error", zap.Error(err))
			return true
		}
	case constants.MessageKindBrain:
		if err := b.HandleBrain(message.Action); err != nil {
			logger.Error("handle brain message error", zap.Error(err))
			return true
		}
	default:
		// logger unsupported message kind
		logger.Error("message error", zap.Error(errors.ErrUnsupportedMessageKind(message.Kind)))
		return true
	}

	// 3. 重新计算 brain 状态并刷新
	if message.Kind != constants.MessageKindBrain {
		b.RefreshState()
	}

	b.queue.Forget(msg)
	return true
}

func (b *BrainLocal) trigLinks(linkIDs ...string) error {
	if len(linkIDs) == 0 {
		return nil
	}

	if err := b.ensureMemoryInit(); err != nil {
		// TODO wrap error
		return err
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

	return nil
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
		b.SendMessage(constants.Message{
			Kind:   constants.MessageKindLink,
			Action: constants.MessageActionLinkReady,
			ID:     link.id,
		})
	}

	return
}

func (b *BrainLocal) RefreshState() {
	inhibitCnt, activateCnt := b.getNeuronCountByState()
	initCnt, waitCnt, readyCnt := b.getLinkCountByState()
	b.logger.Debug("count link and neuron by state",
		zap.Int("neuron_inhibit", inhibitCnt), zap.Int("neuron_activated", activateCnt),
		zap.Int("link_init", initCnt), zap.Int("link_wait", waitCnt), zap.Int("link_ready", readyCnt))
	// send brain sleep message
	if activateCnt+waitCnt+readyCnt == 0 {
		b.SendMessage(constants.Message{
			Kind:   constants.MessageKindBrain,
			Action: constants.MessageActionBrainSleep,
		})
	} else { // > 0, set to running
		b.state = BrainStateRunning
	}
}

func (b *BrainLocal) getNeuronCountByState() (int, int) {
	var inhibitCnt, activateCnt int
	for _, neuron := range b.neurons {
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
	for _, link := range b.links {
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

func (b *BrainLocal) HandleLink(action constants.MessageAction, linkID string) error {
	link := b.GetLink(linkID)
	if link == nil {
		return errors.ErrLinkNotFound(linkID)
	}

	switch action {
	case constants.MessageActionLinkInit:
		// do nothing
	case constants.MessageActionLinkWait:
		// do nothing
	case constants.MessageActionLinkReady:
		destNeuron := b.GetNeuron(link.to)
		if destNeuron == nil {
			return errors.ErrNeuronNotFound(link.to)
		}
		// try dest neuron activate
		b.SendMessage(constants.Message{
			Kind:   constants.MessageKindNeuron,
			Action: constants.MessageActionNeuronTryActivate,
			ID:     destNeuron.id,
		})

	default:
		return errors.ErrUnsupportedMessageAction(action)
	}

	return nil
}

func (b *BrainLocal) HandleNeuron(action constants.MessageAction, neuronID string) error {
	neuron := b.GetNeuron(neuronID)
	if neuron == nil {
		return errors.ErrNeuronNotFound(neuronID)
	}

	switch action {
	case constants.MessageActionNeuronTryInhibit:
		// do nothing , neuron 只有自己 Inhibit
	case constants.MessageActionNeuronTryActivate:
		return b.tryActivateNeuron(neuron)
	default:
		return errors.ErrUnsupportedMessageAction(action)
	}

	return nil
}

func (b *BrainLocal) HandleBrain(action constants.MessageAction) error {
	switch action {
	case constants.MessageActionBrainSleep:
		b.ShutDown()
		b.allLinkNeuronInhibit()
		b.state = BrainStateSleeping
		return nil
	default:
		return errors.ErrUnsupportedMessageAction(action)
	}
}

func (b *BrainLocal) allLinkNeuronInhibit() {
	for _, link := range b.links {
		link.state = LinkStateInit
	}
	for _, neuron := range b.neurons {
		neuron.state = NeuronStateInhibited
	}
}

func (b *BrainLocal) tryActivateNeuron(neuron *Neuron) error {
	if neuron.state == NeuronStateActivated {
		b.logger.Debug("neuron already activated", zap.String("neuron", neuron.id))
		return nil
	}

	should := b.ifNeuronShouldActivate(neuron)
	if !should {
		b.logger.Debug("neuron should not be activated", zap.String("neuron", neuron.id))
		return nil
	}

	// should END, send brain sleep message
	if neuron.id == EndNeuronID {
		b.logger.Info("arrival at END neuron")
		b.SendMessage(constants.Message{
			Kind:   constants.MessageKindBrain,
			Action: constants.MessageActionBrainSleep,
		})
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
	for _, group := range neuron.castGroups {
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
	// 选中的 cast group 中的 link 状态为 wait 的设置为 ready，SendMessage （为 init 的则不改变）
	for linkID, _ := range neuron.castGroups[selected] {
		link := b.GetLink(linkID)
		if link.state == LinkStateWait {
			link.state = LinkStateReady
			b.SendMessage(constants.Message{
				Kind:   constants.MessageKindLink,
				Action: constants.MessageActionLinkReady,
				ID:     link.id,
			})
		}
	}

	return nil
}

func (b *BrainLocal) ensureMemoryInit() error {
	if b.memory != nil {
		return nil
	}

	return b.initMemory()
}

func (b *BrainLocal) initMemory() error {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: b.memoryNumCounters, // number of keys to track frequency of (10M).
		MaxCost:     b.memoryMaxCost,     // maximum cost of cache (1GB).
		BufferItems: 64,                  // number of keys per Get buffer.
	})
	if err != nil {
		// TODO Wrap error
		return err
	}
	b.memory = cache

	return nil
}
