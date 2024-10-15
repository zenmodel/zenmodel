package brainlite

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/core"
	"github.com/zenmodel/zenmodel/internal/utils"
	"github.com/zenmodel/zenmodel/internal/errors"
)

const (
	//  length of brain event queue
	bQueueLen = 10
	// default length of neuron process queue
	defaultNQueueLen = 10
	// default number of neuron process workers
	defaultNWorkerNum = 4
)

func BuildBrain(blueprint core.MultiLangBlueprint, withOpts ...Option) *BrainLite {
	b := &BrainLite{
		id:      utils.GenID(),
		labels:  utils.LabelsDeepCopy(blueprint.GetLabels()),
		state:   core.BrainStateShutdown,
		neurons: make(map[string]*neuron),
		links:   make(map[string]*link),
	}
	b.cond = sync.NewCond(&b.mu)

	for _, l := range blueprint.ListLinks() {
		lk := newLink(l)
		b.links[lk.id] = lk
	}
	for _, n := range blueprint.ListNeurons() {
		neu := newNeuron(n, b.links)
		b.neurons[neu.id] = neu
	}

	// init config
	b.logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			var c string
			if cc, ok := i.(string); ok {
				c = cc
			}
			if len(c) > 0 && strings.Contains(c, "/") {
				lastIndex := strings.LastIndex(c, "/")
				left := c[:lastIndex]
				c = c[lastIndex+1:]
				if strings.Contains(left, "/") {
					lastIndex = strings.LastIndex(left, "/")
					c = left[lastIndex+1:] + "/" + c
				}
			}
			return c
		},
	}).With().Caller().Timestamp().Logger().Level(zerolog.InfoLevel)
	b.BrainMaintainer.nQueueLen = defaultNQueueLen
	b.BrainMaintainer.nWorkerNum = defaultNWorkerNum
	b.BrainMemory.datasourceName = fmt.Sprintf("%s.db", b.id)

	for _, opt := range withOpts {
		opt.apply(b)
	}

	b.logger = b.logger.With().Str("brainID", b.id).Logger()

	b.logger.Info().Interface("blueprint", blueprint).Msg("brain build success")
	return b
}

type BrainLite struct {
	id     string
	labels map[string]string

	neurons map[string]*neuron
	links   map[string]*link

	// brain is in the Running state when there are 1 or more Activate neuron or 1 or more StandBy link.
	state core.BrainState
	// brain memories
	BrainMemory
	BrainMaintainer

	logger zerolog.Logger
	mu     sync.Mutex
	cond   *sync.Cond
}

type BrainMaintainer struct {
	bQueue chan maintainEvent
	stop   chan struct{}

	NeuronRunner
}

type NeuronRunner struct {
	nQueue     chan string
	nQueueLen  int
	nWorkerNum int
}

func (b *BrainLite) TrigLinks(links ...core.Link) error {
	linkIDs := make([]string, 0)
	for _, l := range links {
		if l == nil || l.GetID() == "" {
			continue
		}
		linkIDs = append(linkIDs, l.GetID())
	}
	return b.trigLinks(linkIDs...)
}

func (b *BrainLite) Entry() error {
	// get all entry links
	linkIDs := make([]string, 0)
	for _, l := range b.links {
		if l.isEntryLink() {
			linkIDs = append(linkIDs, l.id)
		}
	}

	return b.trigLinks(linkIDs...)
}

func (b *BrainLite) EntryWithMemory(keysAndValues ...interface{}) error {
	if err := b.SetMemory(keysAndValues...); err != nil {
		return err
	}

	return b.Entry()
}

func (b *BrainLite) SetMemory(keysAndValues ...interface{}) error {
	if len(keysAndValues)%2 != 0 {
		return fmt.Errorf("key and value are not paired")
	}
	if err := b.ensureMemoryInit(); err != nil {
		return err
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		// TODO batch set
		if err := b.BrainMemory.Set(k, v); err != nil {
			return errors.Wrapf(err, "set memory failed")
		}
		b.logger.Debug().
			Any("key", k).
			Any("value", v).
			Msg("set memory")
	}

	return nil
}

func (b *BrainLite) GetMemory(key any) any {
	if b.BrainMemory.db == nil {
		return nil
	}
	v, err := b.BrainMemory.Get(key)
	if err != nil {
		b.logger.Error().Err(err).Msg("get memory failed")
		return nil
	}

	return v
}

func (b *BrainLite) ExistMemory(key any) bool {
	if b.BrainMemory.db == nil {
		return false
	}

	_, err := b.BrainMemory.Get(key)
	if err != nil {
		b.logger.Error().Err(err).Msg("get memory failed")
		return false
	}

	return true
}

func (b *BrainLite) DeleteMemory(key any) {
	if b.BrainMemory.db == nil {
		return
	}

	if err := b.BrainMemory.Del(key); err != nil {
		b.logger.Error().Err(err).Msg("delete memory failed")
	}
}

func (b *BrainLite) ClearMemory() {
	if b.BrainMemory.db == nil {
		return
	}

	if err := b.BrainMemory.Clear(); err != nil {
		b.logger.Error().Err(err).Msg("clear memory failed")
	}
}

func (b *BrainLite) GetState() core.BrainState {
	return b.getState()
}

func (b *BrainLite) Wait() {
	// block when brain running
	b.mu.Lock()
	for b.state != core.BrainStateSleeping && b.state != core.BrainStateShutdown {
		b.cond.Wait()
	}
	b.mu.Unlock()
}

func (b *BrainLite) Shutdown() {
	b.logger.Info().Msg("brain local shutdown")
	close(b.BrainMaintainer.nQueue)
	close(b.BrainMaintainer.bQueue)
	if err := b.BrainMemory.Close(); err != nil {
		b.logger.Error().Err(err).Msg("close memory failed")
	}
	b.setState(core.BrainStateShutdown)
}

func (b *BrainLite) trigLinks(linkIDs ...string) error {
	if len(linkIDs) == 0 {
		return nil
	}

	if err := b.ensureMemoryInit(); err != nil {
		// TODO wrap error
		return err
	}

	// ensure brain maintainer start
	b.ensureMaintainerStart()

	// goroutine wait
	var wg sync.WaitGroup
	for _, linkID := range linkIDs {
		l, ok := b.links[linkID]
		if !ok {
			continue
		}

		wg.Add(1)
		go b.trigLink(&wg, l)
	}
	wg.Wait()

	b.refreshState()

	return nil
}

func (b *BrainLite) trigLink(wg *sync.WaitGroup, l *link) {
	defer wg.Done()

	if l.status.state != core.LinkStateReady {
		// change link state as ready
		l.status.state = core.LinkStateReady

		// send maintain event
		b.publishEvent(maintainEvent{
			kind:   eventKindLink,
			action: eventActionLinkReady,
			id:     l.id,
		})
	}

	return
}

func (b *BrainLite) ensureMemoryInit() error {
	if b.BrainMemory.db != nil {
		return nil
	}

	return b.BrainMemory.Init()
}