package pipelines

type config struct {
	logger   Logger
	stations []*stationConfig
}
type option func(*config)

var Options singleton

type singleton struct{}

func (singleton) Logger(logger Logger) option {
	return func(c *config) { c.logger = logger }
}
func (singleton) StationSingleton(station Station) option {
	return Options.StationFactory(func() Station { return station })
}
func (singleton) StationFactory(stationFunc func() Station) option {
	return func(c *config) {
		c.stations = append(c.stations, &stationConfig{stationFunc: stationFunc})
	}
}
func (singleton) WorkerCount(count int) option {
	return func(c *config) { c.stations[len(c.stations)-1].workerCount = count }
}

func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.Logger(nop{}),
	}, options...)
}

type nop struct{}

func (nop) Printf(_ string, _ ...interface{}) {}
