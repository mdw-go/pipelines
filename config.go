package pipelines

type config struct {
	logger Logger
	groups []*group
}

func (this *config) apply(options ...option) {
	for _, option := range append(Options.defaults(), options...) {
		if option != nil {
			option(this)
		}
	}
}

type option func(*config)

var Options singleton

type singleton struct{}

func (singleton) Logger(logger Logger) option {
	return func(c *config) { c.logger = logger }
}
func (singleton) StationGroup(stations ...Station) option {
	if len(stations) == 0 {
		return nil
	}
	return func(c *config) { c.groups = append(c.groups, &group{stations: stations}) }
}
func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.Logger(nop{}),
	}, options...)
}

type nop struct{}

func (nop) Printf(_ string, _ ...any) {}
