package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func isEvenValue(i int64) bool {
	return i%2 == 0
}

func by2(i int64) int64 {
	return i * 2
}

func printFilterResult(i int64) {
	fmt.Printf("Result: %d \n", i)
}

// Example4 demonstrates how to utilize error handling in a pipeline and use steps as filters
func Example4() {

	builder := &pip.Builder[int64]{}

	// the filter step
	step1 := builder.NewStep(pip.StepFilterConfig[int64]{
		Label:        "step1",
		Replicas:     1,
		PassCriteria: isEvenValue,
	})

	// The processing step
	step2 := builder.NewStep(pip.StepBasicConfig[int64]{
		Label:    "step2",
		Replicas: 1,
		Process:  by2,
	})

	// The result step
	resultStep := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "resultStep",
		Replicas: 1,
		Process:  printFilterResult,
	})

	// init pipeline
	pConfig := pip.PipelineConfig{
		DefaultStepChannelSize: 10,
		TrackTokensCount:       true,
	}
	pipeline := builder.NewPipeline(pConfig, step1, step2, resultStep)
	pipeline.Init()

	// Running
	ctx := context.Background()
	go pipeline.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pipeline.FeedOne(i)
	}

	// waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 4 Done!!!")
}
