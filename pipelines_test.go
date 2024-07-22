package pipelines_test

import (
	"sync/atomic"
	"testing"

	"github.com/mdwhatcott/pipelines"
)

func Test(t *testing.T) {
	input := make(chan any)
	go func() {
		defer close(input)
		for x := range 50 {
			input <- x + 1
		}
	}()

	sum := new(atomic.Int64)
	listener := pipelines.New(input,
		pipelines.Options.Logger(TLogger{T: t}),
		pipelines.Options.StationFactory(NewSquares),
		pipelines.Options.StationSingleton(NewFirstN(20)),
		pipelines.Options.StationFactory(NewEvens), pipelines.Options.FanOut(5),
		pipelines.Options.StationFactory(NewDuplicate), pipelines.Options.FanOut(5),
		pipelines.Options.StationSingleton(NewSum(sum)), pipelines.Options.FanOut(5),
	)

	listener.Listen()

	if total := sum.Load(); total != 3080 {
		t.Error("Expected 3080, got:", total)
	}
}

type TLogger struct{ *testing.T }

func (this TLogger) Printf(format string, args ...any) {
	this.Logf(format, args...)
}

///////////////////////////////

type Squares struct{}

func NewSquares() pipelines.Station {
	return &Squares{}
}

func (this *Squares) Do(input any, output func(any)) {
	switch input := input.(type) {
	case int:
		output(input * input)
	}
}

///////////////////////////////

type Evens struct{}

func NewEvens() pipelines.Station {
	return &Evens{}
}

func (this *Evens) Do(input any, output func(any)) {
	switch input := input.(type) {
	case int:
		if input%2 == 0 {
			output(input)
		}
	}
}

///////////////////////////////

type FirstN struct {
	N       *atomic.Int64
	handled *atomic.Int64
}

func NewFirstN(n int64) pipelines.Station {
	N := new(atomic.Int64)
	N.Add(n)
	return &FirstN{N: N, handled: new(atomic.Int64)}
}

func (this *FirstN) Do(input any, output func(any)) {
	if this.handled.Load() >= this.N.Load() {
		return
	}
	output(input)
	this.handled.Add(1)
}

///////////////////////////////

type Duplicate struct{}

func NewDuplicate() pipelines.Station {
	return &Duplicate{}
}

func (this *Duplicate) Do(input any, output func(any)) {
	output(input)
	output(input)
}

///////////////////////////////

type Sum struct {
	sum *atomic.Int64
}

func NewSum(sum *atomic.Int64) pipelines.Station {
	return &Sum{sum: sum}
}

func (this *Sum) Do(input any, output func(any)) {
	switch input := input.(type) {
	case int:
		this.sum.Add(int64(input))
		output(input)
	}
}
