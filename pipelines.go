package pipelines

import "sync"

func New(input chan any, options ...option) Listener {
	config := new(config)
	config.apply(options...)
	return &listener{
		input:    input,
		logger:   config.logger,
		stations: config.stations,
	}
}

type listener struct {
	logger   Logger
	stations []*stationConfig
	input    chan any
}

func (this *listener) Listen() {
	input := this.input
	for _, station := range this.stations {
		output := make(chan any)
		go station.run(input, output)
		input = output
	}
	for v := range input {
		this.logger.Printf("value at end of pipeline: %v", v)
	}
}

type stationConfig struct {
	stationFunc func() Station
	workerCount int
}

func (this *stationConfig) run(input, output chan any) {
	if this.workerCount > 1 {
		this.runFannedOutStation(input, output)
	} else {
		this.runStation(input, output)
	}
}

func (this *stationConfig) runFannedOutStation(input, final chan any) {
	defer close(final)
	var outs []chan any
	for range this.workerCount {
		out := make(chan any)
		outs = append(outs, out)
		go this.runStation(input, out)
	}
	var waiter sync.WaitGroup
	waiter.Add(len(outs))
	defer waiter.Wait()
	for _, out := range outs {
		go func(out chan any) {
			defer waiter.Done()
			for item := range out {
				final <- item
			}
		}(out)
	}
}
func (this *stationConfig) runStation(input, output chan any) {
	defer close(output)
	action := this.stationFunc()
	for input := range input {
		action.Do(input, func(v any) { output <- v })
	}
}
