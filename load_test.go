package pipelines_test

import (
	"testing"

	"github.com/mdw-go/pipelines"
)

func TestLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var group1 []pipelines.Station
	for range 1024 {
		group1 = append(group1, NewLoadTestStation())
	}

	var group2 []pipelines.Station
	for range 8 {
		group2 = append(group2, NewLoadTestStation())
	}

	input := make(chan any)
	go func() {
		defer close(input)
		for range 10_000_000 {
			input <- struct{}{}
		}
	}()

	listener := pipelines.New(input,
		pipelines.Options.Logger(&TLogger{T: t}),
		pipelines.Options.StationGroup(group1...),
		pipelines.Options.StationGroup(group2...),
		pipelines.Options.StationGroup(NewLoadTestStation()),
		pipelines.Options.StationGroup(NewLoadTestFinalStation(t)),
	)
	listener.Listen()

	for _, station := range append(group1, group2...) {
		if station.(*LoadTestStation).count == 0 {
			t.Error("a fanned-out station handled 0 items")
		}
	}
}

/////////////////////////////

type LoadTestStation struct {
	count int
}

func NewLoadTestStation() *LoadTestStation {
	return &LoadTestStation{}
}

func (this *LoadTestStation) Do(input any, output func(any)) {
	this.count++
	output(input)
}

//////////////////////////////////

type LoadTestFinalStation struct {
	t     *testing.T
	count int
}

func NewLoadTestFinalStation(t *testing.T) *LoadTestFinalStation {
	return &LoadTestFinalStation{t: t}
}
func (this *LoadTestFinalStation) Do(input any, output func(any)) {
	this.count++
	if this.count%100_000 == 0 {
		this.t.Logf("progress: %d", this.count)
	}
}
func (this *LoadTestFinalStation) Finalize(_ func(any)) {
	this.t.Logf("Station finished after processing %d items", this.count)
}
