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

	SetMemory(keysAndValues ...interface{}) error
	GetMemory(key interface{}) (interface{}, bool)
	DeleteMemory(key interface{})

	// SetMemoryStream()
	// WatchMemoryStream()
}
