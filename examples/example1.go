package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func plus5(i int64) int64 {
	return i + 5
}
func minus10(i int64) int64 {
	return i - 10
}
func printResult(i int64) {
	fmt.Printf("Result: %d \n", i)
}

func Example1() {

	builder := &pip.Builder[int64]{}

	plus5Step := builder.NewStep(pip.StepBasicConfig[int64]{
		Label:    "plus5",
		Replicas: 1,
		Process:  plus5,
	})
	minus10Step := builder.NewStep(pip.StepBasicConfig[int64]{
		Label:    "minus10",
		Replicas: 1,
		Process:  minus10,
	})
	printResultStep := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "print",
		Replicas: 1,
		Process:  printResult,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepInputChannelSize: 10,
		TrackTokensCount:            true,
	}
	pipeline := builder.NewPipeline(pConfig, plus5Step, minus10Step, printResultStep) // Notice 10 is the buffer size
	pipeline.Init()

	// Running
	ctx := context.Background()
	pipeline.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 25; i++ {
		pipeline.FeedOne(int64(i))
	}

	// 	waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 1 Done!!!")
}
