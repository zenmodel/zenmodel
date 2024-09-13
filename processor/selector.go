package processor

const (
	DefaultCastGroupName = "__DEFAULT_CAST_GROUP__"
)

type Selector interface {
	// Select 选择传播组
	Select(ctx BrainContextReader) string
	// Clone 克隆当前 Selector,区别于 DeepCopy() 运行时的状态不应该拷贝
	Clone() Selector
}

type DefaultSelector struct{}

func (s *DefaultSelector) Select(ctx BrainContextReader) string {
	return DefaultCastGroupName
}

func (s *DefaultSelector) Clone() Selector {
	return &DefaultSelector{}
}

func NewFuncSelector(selectFn func(ctx BrainContextReader) string) *FuncSelector {
	return &FuncSelector{
		selectFn: selectFn,
	}
}

type FuncSelector struct {
	selectFn func(ctx BrainContextReader) string
}

func (s *FuncSelector) Select(ctx BrainContextReader) string {
	return s.selectFn(ctx)
}

func (s *FuncSelector) Clone() Selector {
	return &FuncSelector{
		selectFn: s.selectFn,
	}
}
