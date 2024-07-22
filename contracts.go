package pipelines

type Logger interface {
	Printf(format string, args ...any)
}

type Listener interface {
	Listen()
}

type Station interface {
	Do(input any, output []any) (n int)
}
