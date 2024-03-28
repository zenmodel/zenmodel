package zenmodel

const (
	BrainStateAwake    BrainState = "Awake"
	BrainStateRunning  BrainState = "Running"
	BrainStateSleeping BrainState = "Sleeping"
)

type Brain interface {
	// TrigLinks 触发指定 Links
	TrigLinks(linkIDs ...string) error
	// Entry 触发所有 Entry Links
	Entry() error
	// EntryWithMemory 先设置 Memory 再触发所有 Entry Links
	EntryWithMemory(keysAndValues ...interface{}) error

	// SetMemory set memories for brain, one key value pair is one memory.
	// memory will lazy initial util `SetMemory` or any link trig
	SetMemory(keysAndValues ...interface{}) error
	// GetMemory get memory by key
	GetMemory(key interface{}) interface{}
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key interface{}) bool
	// DeleteMemory delete one memory by key
	DeleteMemory(key interface{})
	// ClearMemory clear all memories
	ClearMemory()
	// GetState get brain state
	GetState() BrainState
	// Wait wait util brain maintainer shutdown, which means brain state is `Sleeping`
	Wait()
}
