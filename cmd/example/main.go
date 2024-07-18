package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mdwhatcott/pipelines"
)

func main() {
	caboose := &Caboose{}
	pipeline := pipelines.New(
		pipelines.WithStation(&SquareStation{}, 5, 1024),
		pipelines.WithStation(&ItoaStation{}, 5, 1024),
		pipelines.WithStation(caboose, 1, 1024),
	)

	for x := range 20 {
		pipeline <- x + 1
	}
	close(pipeline)

	time.Sleep(time.Millisecond * 10)

	fmt.Println(caboose.collected)
}

type ItoaStation struct{}

func (this *ItoaStation) Station(input any, output []any) (result int) {
	switch input := input.(type) {
	case int:
		output[result] = strconv.Itoa(input)
		result++
	default:
		output[result] = input
	}
	return result
}

type SquareStation struct{}

func (this *SquareStation) Station(input any, output []any) (result int) {
	switch input := input.(type) {
	case int:
		output[result] = input * input
		result++
	default:
		output[result] = input
	}
	return result
}

type Caboose struct {
	collected []any
}

func (this *Caboose) Station(input any, _ []any) int {
	this.collected = append(this.collected, input)
	return 0
}
