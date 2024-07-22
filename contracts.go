package pipelines

type Logger interface {
	Printf(format string, args ...interface{})
}

type Listener interface {
	Listen()
}

type Station interface {
	Do(input any, output []any) (n int)
}
