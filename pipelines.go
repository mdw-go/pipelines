package pipelines

import (
	"container/list"
	"sync"
)

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

type stationConfig struct {
	stationFunc func() Station
	workerCount int
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
	outputs := list.New()
	for input := range inputs {
		action.Do(input, outputs)
		for this := outputs.Front(); this != nil; {
			output <- this.Value
			next := this.Next()
			outputs.Remove(this)
			this = next
		}
	}
}
