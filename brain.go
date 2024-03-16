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

	// SetMemory 设置 Memory
	SetMemory(keysAndValues ...interface{}) error
	GetMemory(key interface{}) (interface{}, bool)
	DeleteMemory(key interface{})

	GetMaintainer() Maintainer

	// SetMemoryStream()
	// WatchMemoryStream()
}
