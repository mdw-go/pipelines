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
	output := pipelines.New(input,
		pipelines.Station(NewSquares(), 1, 1024),
		pipelines.Station(NewEvens(), 1, 1024),
		pipelines.Station(NewFirstN(20), 1, 1024),
		pipelines.Station(NewDuplicate(), 1, 1024),
		pipelines.Station(NewSum(sum), 10, 1024),
	)

	for x := range output {
		t.Log(x)
	}

	if total := sum.Load(); total != 22960 {
		t.Error("Expected 22960, got:", total)
	}
}

type Squares struct{}

func NewSquares() *Squares {
	return &Squares{}
}

func (this *Squares) Do(input any, output []any) (n int) {
	switch input := input.(type) {
	case int:
		n = pipelines.Append(output, n, input*input)
	}
	return n
}

type Evens struct{}

func NewEvens() *Evens {
	return &Evens{}
}

func (this *Evens) Do(input any, output []any) (n int) {
	switch input := input.(type) {
	case int:
		if input%2 == 0 {
			n = pipelines.Append(output, n, input)
		}
	}
	return n
}

type FirstN struct {
	N       int
	handled int
}

func NewFirstN(n int64) *FirstN {
	return &FirstN{
		N:       int(n),
		handled: 0,
	}
}

func (this *FirstN) Do(input any, output []any) (n int) {
	if this.handled >= this.N {
		return 0
	}
	switch input := input.(type) {
	case int:
		n = pipelines.Append(output, n, input)
		this.handled++
	}
	return n
}

type Duplicate struct{}

func NewDuplicate() *Duplicate {
	return &Duplicate{}
}

func (this *Duplicate) Do(input any, output []any) (n int) {
	return pipelines.Append(output, n, input, input)
}

type Sum struct {
	sum *atomic.Int64
}

func NewSum(sum *atomic.Int64) *Sum {
	return &Sum{sum: sum}
}

func (this *Sum) Do(input any, outputs []any) (n int) {
	switch input := input.(type) {
	case int:
		this.sum.Add(int64(input))
		n = pipelines.Append(outputs, n, input)
	}
	return n
}
