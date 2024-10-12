package core

const (
	EntryLinkFrom = "__EXTERNAL_SIGNAL__"
	EndLinkTo     = EndNeuronID
)

type LinkState string

const (
	LinkStateInit  LinkState = "Init"
	LinkStateWait  LinkState = "Wait"
	LinkStateReady LinkState = "Ready"
)

type Link interface {
	GetID() string
	GetLabels() map[string]string
	GetSrcNeuronID() string
	GetDestNeuronID() string
	IsEntryLink() bool
	IsEndLink() bool

	SetLabels(labels map[string]string)
}

// LinkOption configures a link.
type LinkOption interface {
	Apply(link Link)
}

// linkOptionFunc wraps a func, so it satisfies the LinkOption interface.
type linkOptionFunc func(Link)

func (f linkOptionFunc) Apply(link Link) {
	f(link)
}

// WithLinkLabels sets the specific labels for Link
func WithLinkLabels(labels map[string]string) LinkOption {
	return linkOptionFunc(func(link Link) {
		link.SetLabels(labels)
	})
}
