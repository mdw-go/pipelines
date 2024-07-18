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
		&SquareStation{},
		&EvenStation{},
		&FirstNStation{N: 20},
		&SumStation{sum: sum},
	)

	for x := range output {
		t.Log(x)
	}

	if total := sum.Load(); total != 11480 {
		t.Error("Expected 11480, got:", total)
	}
}

type SquareStation struct{}

func (this *SquareStation) Do(input any, output []any) (n int) {
	switch input := input.(type) {
	case int:
		output[n] = input * input
		n++
	}
	return n
}

type EvenStation struct{}

func (this *EvenStation) Do(input any, output []any) (n int) {
	switch input := input.(type) {
	case int:
		if input%2 == 0 {
			output[n] = input
			n++
		}
	}
	return n
}

type FirstNStation struct{ N, handled int }

func (this *FirstNStation) Do(input any, output []any) (n int) {
	if this.handled >= this.N {
		return 0
	}
	switch input := input.(type) {
	case int:
		output[n] = input
		this.handled++
		n++
	}
	return n
}

type SumStation struct{ sum *atomic.Int64 }

func (this *SumStation) FanoutCount() int    { return 5 }
func (this *SumStation) MaxOutputCount() int { return 1 }
func (this *SumStation) Do(input any, outputs []any) (n int) {
	switch input := input.(type) {
	case int:
		this.sum.Add(int64(input))
		outputs[n] = input
		n++
	}
	return n
}
