package brainlocal

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/rs/zerolog"
	"github.com/zenmodel/zenmodel/brain"
	"github.com/zenmodel/zenmodel/internal/utils"
)

const (
	//  length of brain event queue
	bQueueLen = 10
	// default length of neuron process queue
	defaultNQueueLen = 10
	// default number of neuron process workers
	defaultNWorkerNum = 4
	// default number of keys to track frequency of (10M)
	defaultMemNumCounters = 1e7
	// default maximum cost of cache (1GB)
	defaultMemMaxCost = 1 << 30
)

func NewBrainLocal(blueprint brain.Blueprint, withOpts ...Option) *BrainLocal {
	b := &BrainLocal{
		id:      utils.GenID(),
		labels:  utils.LabelsDeepCopy(blueprint.GetLabels()),
		state:   brain.BrainStateShutdown,
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
	b.BrainMemory.numCounters = defaultMemNumCounters
	b.BrainMemory.maxCost = defaultMemMaxCost

	for _, opt := range withOpts {
		opt.apply(b)
	}

	b.logger = b.logger.With().Str("brainID", b.id).Logger()

	b.logger.Info().Interface("blueprint", blueprint).Msg("brain build success")
	return b
}

type BrainLocal struct {
	id     string
	labels map[string]string

	neurons map[string]*neuron
	links   map[string]*link

	// brain is in the Running state when there are 1 or more Activate neuron or 1 or more StandBy link.
	state brain.BrainState
	// brain memories
	BrainMemory
	BrainMaintainer

	logger zerolog.Logger
	mu     sync.Mutex
	cond   *sync.Cond
}

type BrainMemory struct {
	cache       *ristretto.Cache
	numCounters int64
	maxCost     int64
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

func (b *BrainLocal) TrigLinks(links ...brain.Link) error {
	linkIDs := make([]string, 0)
	for _, l := range links {
		if l == nil || l.GetID() == "" {
			continue
		}
		linkIDs = append(linkIDs, l.GetID())
	}
	return b.trigLinks(linkIDs...)
}

func (b *BrainLocal) Entry() error {
	// get all entry links
	linkIDs := make([]string, 0)
	for _, l := range b.links {
		if l.isEntryLink() {
			linkIDs = append(linkIDs, l.id)
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
		b.BrainMemory.cache.Set(k, v, 1) // TODO maybe calculate cost
		b.logger.Debug().
			Any("key", k).
			Any("value", v).
			Msg("set memory")
	}
	b.BrainMemory.cache.Wait()

	return nil
}

func (b *BrainLocal) GetMemory(key any) any {
	if b.BrainMemory.cache == nil {
		return nil
	}
	v, _ := b.BrainMemory.cache.Get(key)

	return v
}

func (b *BrainLocal) ExistMemory(key any) bool {
	if b.BrainMemory.cache == nil {
		return false
	}

	_, ok := b.BrainMemory.cache.Get(key)

	return ok
}

func (b *BrainLocal) DeleteMemory(key any) {
	if b.BrainMemory.cache == nil {
		return
	}

	b.BrainMemory.cache.Del(key)
}

func (b *BrainLocal) ClearMemory() {
	if b.BrainMemory.cache == nil {
		return
	}

	b.BrainMemory.cache.Clear()
}

func (b *BrainLocal) GetState() brain.BrainState {
	return b.getState()
}

func (b *BrainLocal) Wait() {
	// block when brain running
	b.mu.Lock()
	for b.state != brain.BrainStateSleeping && b.state != brain.BrainStateShutdown {
		b.cond.Wait()
	}
	b.mu.Unlock()
}

func (b *BrainLocal) Shutdown() {
	b.logger.Info().Msg("brain local shutdown")
	close(b.BrainMaintainer.nQueue)
	close(b.BrainMaintainer.bQueue)
	b.BrainMemory.cache.Close()
	b.setState(brain.BrainStateShutdown)
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

func (b *BrainLocal) trigLink(wg *sync.WaitGroup, l *link) {
	defer wg.Done()

	if l.status.state != brain.LinkStateReady {
		// change link state as ready
		l.status.state = brain.LinkStateReady

		// send maintain event
		b.publishEvent(maintainEvent{
			kind:   eventKindLink,
			action: eventActionLinkReady,
			id:     l.id,
		})
	}

	return
}

func (b *BrainLocal) ensureMemoryInit() error {
	if b.BrainMemory.cache != nil {
		return nil
	}

	return b.initMemory()
}

func (b *BrainLocal) initMemory() error {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: b.BrainMemory.numCounters,
		MaxCost:     b.BrainMemory.maxCost,
		BufferItems: 64, // number of keys per Get buffer.
	})
	if err != nil {
		// TODO Wrap error
		return err
	}
	b.BrainMemory.cache = cache

	return nil
}
