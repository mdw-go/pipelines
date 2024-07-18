package pipelines

import "sync"

func New(input chan any, actions ...action) chan any {
	for _, action := range actions {
		output := make(chan any)
		go fanout(input, output, action)
		input = output
	}
	return input // which is now the final output
}

type action interface {
	Do(input any, output []any) (n int)
}
type fanoutCount interface {
	FanoutCount() int
}
type maxOutputCount interface {
	MaxOutputCount() int
}

func fanout(input, final chan any, action action) {
	count, ok := action.(fanoutCount)
	if !ok {
		station(input, final, action)
		return
	}
	defer close(final)
	var outs []chan any
	for range max(1, count.FanoutCount()) {
		out := make(chan any)
		outs = append(outs, out)
		go station(input, out, action)
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
func station(inputs, output chan any, action action) {
	defer close(output)
	maxOutput, ok := action.(maxOutputCount)
	maxOutputCount := 64
	if ok {
		maxOutputCount = max(1, maxOutput.MaxOutputCount())
	}
	outputs := make([]any, maxOutputCount)
	for input := range inputs {
		n := action.Do(input, outputs)
		for o := range n {
			output <- outputs[o]
			outputs[o] = nil
		}
	}
}
