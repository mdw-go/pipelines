package pipelines

import "container/list"

type Logger interface {
	Printf(format string, args ...any)
}

type Listener interface {
	Listen()
}

type Station interface {
	Do(input any, output *list.List)
}

type Output interface {
	PushBack(any)
}
