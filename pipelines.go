package pipelines

import "sync"

func New(input chan any, options ...option) Listener {
	config := new(config)
	for _, option := range append(Options.defaults(), options...) {
		option(config)
	}
	return &listener{logger: config.logger, input: input, stations: config.stations}
}

type listener struct {
	logger   Logger
	stations []*stationConfig
	input    chan any
}

func (this *listener) Listen() {
	input := this.input
	for _, config := range this.stations {
		output := make(chan any)
		if config.workerCount > 1 {
			go runFannedOutStation(input, output, config)
		} else {
			go runStation(input, output, config)
		}
		input = output
	}
	for v := range input {
		this.logger.Printf("unhandled value at end of pipeline: %v", v)
	}
}

func StationFunc[S any](v S) func() Station {
	return func() Station { return any(v).(Station) }
}

func Append(outputs []any, n int, vs ...any) int {
	for _, v := range vs {
		outputs[n] = v
		n++
	}
	return n
}

type stationConfig struct {
	stationFunc      func() Station
	workerCount      int
	outputBufferSize int
}

func runFannedOutStation(input, final chan any, config *stationConfig) {
	defer close(final)
	var outs []chan any
	for range config.workerCount {
		out := make(chan any)
		outs = append(outs, out)
		go runStation(input, out, config)
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
func runStation(inputs, output chan any, config *stationConfig) {
	defer close(output)
	action := config.stationFunc()
	outputs := make([]any, config.outputBufferSize)
	for input := range inputs {
		n := action.Do(input, outputs)
		for o := range n {
			output <- outputs[o]
			outputs[o] = nil
		}
	}
}
