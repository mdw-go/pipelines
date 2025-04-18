package pipelines

import "sync"

func New(input chan any, options ...option) Listener {
	config := new(config)
	config.apply(options...)
	return &listener{
		input:  input,
		logger: config.logger,
		groups: config.groups,
	}
}

type listener struct {
	logger Logger
	groups []*group
	input  chan any
}

func (this *listener) Listen() {
	input := this.input
	for _, group := range this.groups {
		output := make(chan any)
		go group.run(input, output)
		input = output
	}
	for v := range input {
		this.logger.Printf("value at end of pipeline: %v", v)
	}
}

type group struct {
	stations []Station
}

func (this *group) run(input, output chan any) {
	if len(this.stations) > 1 {
		this.runFannedOutStation(input, output)
	} else {
		this.runStation(this.stations[0], input, output)
	}
}
func (this *group) runFannedOutStation(input, final chan any) {
	defer close(final)
	var outs []chan any
	for _, station := range this.stations {
		out := make(chan any)
		outs = append(outs, out)
		go this.runStation(station, input, out)
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
func (this *group) runStation(station Station, input, output chan any) {
	defer close(output)
	out := func(v any) { output <- v }
	if finalizer, ok := station.(Finalizer); ok {
		defer finalizer.Finalize(out)
	}
	for input := range input {
		station.Do(input, out)
	}
}
