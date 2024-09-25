package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func plus5(i int64) (int64, error) {
	return i + 5, nil
}
func minus10(i int64) (int64, error) {
	return i - 10, nil
}
func printResult(i int64) error {
	fmt.Printf("Result: %d \n", i)
	return nil
}

func Example1() {

	builder := &pip.Builder[int64]{}

	plus5Step := builder.NewStep(&pip.StepConfig[int64]{
		Label:    "plus5",
		Replicas: 1,
		Process:  plus5,
	})
	minus10Step := builder.NewStep(&pip.StepConfig[int64]{
		Label:    "minus10",
		Replicas: 1,
		Process:  minus10,
	})
	printResultStep := builder.NewStep(&pip.StepResultConfig[int64]{
		Label:    "print",
		Replicas: 1,
		Process:  printResult,
	})

	pipeline := builder.NewPipeline(10, plus5Step, minus10Step, printResultStep) // Notice 10 is the buffer size
	pipeline.Init()

	// Running
	ctx := context.Background()
	pipeline.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 500; i++ {
		pipeline.FeedOne(int64(i))
	}

	// 	waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 1 Done!!!")
}
