package pipelines_test

import (
	"errors"
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
	closed := new(atomic.Int64)
	logger := TLogger{T: t}
	fanout := 5
	catchAll := NewErrorCatch()
	listener := pipelines.New(input,
		pipelines.Options.Logger(logger),
		pipelines.Options.StationFactory(NewSquares),
		pipelines.Options.StationFactory(NewEvens),
		pipelines.Options.StationSingleton(NewFirstN(10)),
		pipelines.Options.StationSingleton(NewSum(sum, closed)), pipelines.Options.FanOut(fanout),
		pipelines.Options.StationSingleton(catchAll),
	)

	listener.Listen()

	const expected = 1140
	if total := sum.Load(); total != expected {
		t.Errorf("Expected %d, got %d", expected, total)
	}
	if final := int(closed.Load()); final != fanout {
		t.Errorf("Expected %d, got %d", fanout, final)
	}
	if len(catchAll.errors) != fanout {
		t.Errorf("Expected %d, got %d", fanout, len(catchAll.errors))
	}
}

type TLogger struct{ *testing.T }

func (this TLogger) Printf(format string, args ...any) {
	this.Helper()
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
	sum       *atomic.Int64
	finalized *atomic.Int64
}

func NewSum(sum, finalized *atomic.Int64) pipelines.Station {
	return &Sum{sum: sum, finalized: finalized}
}

func (this *Sum) Do(input any, output func(any)) {
	switch input := input.(type) {
	case int:
		this.sum.Add(int64(input))
		output(input)
	}
}

func (this *Sum) Close() error {
	this.finalized.Add(1)
	return errors.New("boink")
}

///////////////////////////////

type ErrorCatch struct {
	errors []error
}

func NewErrorCatch() *ErrorCatch {
	return &ErrorCatch{}
}

func (this *ErrorCatch) Do(input any, output func(any)) {
	switch input := input.(type) {
	case error:
		this.errors = append(this.errors, input)
	}
}
