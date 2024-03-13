package zenmodel

const (
	BrainStateAwake    BrainState = "Awake"
	BrainStateRunning  BrainState = "Running"
	BrainStateSleeping BrainState = "Sleeping"
)

type Brain interface {
	// TrigLinks 触发指定 Links
	TrigLinks(linkIDs ...string)
	// Entry 触发所有 Entry Links
	Entry()

	//SetContext()
	//SetContextUnsafe()
	//SetContextField()
	//SetContextFieldunsafe()
	//GetContext()
	//GetOutput() // 如果是 stream , 则是 stream channel 中的元素拼接而成的结果
	//WatchOutput()
	//GetStatus()
}
