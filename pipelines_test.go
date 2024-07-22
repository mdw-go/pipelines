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
		pipelines.Station(&Squares{}, 1, 1024),
		pipelines.Station(&Evens{}, 1, 1024),
		pipelines.Station(&FirstN{N: 20}, 1, 1024),
		pipelines.Station(&Sum{sum: sum}, 10, 1024),
	)

	for x := range output {
		t.Log(x)
	}

	if total := sum.Load(); total != 11480 {
		t.Error("Expected 11480, got:", total)
	}
}

type Squares struct{}

func (this *Squares) Do(input any, output []any) (n int) {
	switch input := input.(type) {
	case int:
		n = pipelines.Append(output, n, input*input)
	}
	return n
}

type Evens struct{}

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

type Sum struct {
	sum *atomic.Int64
}

func (this *Sum) Do(input any, outputs []any) (n int) {
	switch input := input.(type) {
	case int:
		this.sum.Add(int64(input))
		n = pipelines.Append(outputs, n, input)
	}
	return n
}
