package core

const (
	// BrainStateShutdown brain 实现所使用的资源均已经释放或清空
	BrainStateShutdown BrainState = "Shutdown"
	// BrainStateSleeping brain 所有 neurons 和 links 均处于不活跃的状态
	BrainStateSleeping BrainState = "Sleeping"
	// BrainStateRunning brain 处于正常运行状态
	BrainStateRunning BrainState = "Running"
)

type BrainState string

type Brain interface {
	// TrigLinks 触发指定 Links
	TrigLinks(links ...Link) error
	// Entry 触发所有 Entry Links
	Entry() error
	// EntryWithMemory 先设置 Memory 再触发所有 Entry Links
	EntryWithMemory(keysAndValues ...any) error

	// SetMemory set memories for brain, one key value pair is one memory.
	// memory will lazy initial util `SetMemory` or any link trig
	SetMemory(keysAndValues ...any) error
	// GetMemory get memory by key
	GetMemory(key any) any
	// ExistMemory indicates whether there is a memory in the brain
	ExistMemory(key any) bool
	// DeleteMemory delete one memory by key
	DeleteMemory(key any)
	// ClearMemory clear all memories
	ClearMemory()
	// GetState get brain state
	GetState() BrainState
	// Wait wait util brain maintainer shutdown, which means brain state is `Sleeping`
	Wait()
	// Shutdown the brain
	Shutdown()
}
