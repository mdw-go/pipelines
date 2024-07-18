package pipelines

import "sync"

type stationConfig struct {
	action           Action
	workerCount      int
	outputBufferSize int
}

func New(configs ...stationConfig) (result chan any) {
	input := make(chan any)
	result = input
	for _, config := range configs {
		output := make(chan any)
		go fanout(input, output, config)
		input = output
	}
	return result
}

type Action interface {
	Station(input any, output []any) (n int)
}

func WithStation(station Action, workerCount int, outputBufferSize int) stationConfig {
	return stationConfig{
		action:           station,
		workerCount:      max(1, min(32, workerCount)),
		outputBufferSize: max(1, min(1024, outputBufferSize)),
	}
}

func fanout(input, final chan any, config stationConfig) {
	defer close(final)
	var outs []chan any
	for range config.workerCount {
		out := make(chan any)
		outs = append(outs, out)
		go station(input, out, config)
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
func station(inputs, output chan any, config stationConfig) {
	defer close(output)
	outputs := make([]any, config.outputBufferSize)
	for input := range inputs {
		n := config.action.Station(input, outputs)
		for o := range n {
			output <- outputs[o]
			outputs[o] = nil
		}
	}
}
