package pipelines_test

import (
	"sync/atomic"
	"testing"

	"github.com/mdwhatcott/pipelines"
)

// Test a somewhat interesting pipeline example, based on this Clojure threading macro example:
// https://clojuredocs.org/clojure.core/-%3E%3E#example-542692c8c026201cdc326a52
// (->> (range) (map #(* % %)) (filter even?) (take 10) (reduce +))  ; output: 1140
func Test(t *testing.T) {
	input := make(chan any)
	go func() {
		defer close(input)
		for x := range 50 {
			input <- x
		}
	}()

	sum := new(atomic.Int64)
	listener := pipelines.New(input,
		pipelines.Options.Logger(TLogger{T: t}),
		pipelines.Options.StationFactory(NewSquares),
		pipelines.Options.StationFactory(NewEvens),
		pipelines.Options.StationSingleton(NewFirstN(10)),
		pipelines.Options.StationSingleton(NewSum(sum)), pipelines.Options.FanOut(5),
	)

	listener.Listen()

	const expected = 1140
	if total := sum.Load(); total != expected {
		t.Errorf("Expected %d, got %d", expected, total)
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
