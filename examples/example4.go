package examples

import (
	"context"
	"fmt"

	"github.com/m-faried/pipelines"
)

func filterOdd(i int64) (int64, error) {
	if i%2 == 0 {
		return i, nil
	}
	return 0, fmt.Errorf("odd number: %d", i)
}

func by2(i int64) (int64, error) {
	return i * 2, nil
}

func printFilterResult(i int64) error {
	fmt.Printf("Result: %d \n", i)
	return nil
}

func filterErrorHandler(id string, err error) {
	fmt.Printf("Error in %s: %s\n", id, err.Error())
}

// Example4 demonstrates how to utilize error handling in a pipeline and use steps as filters
func Example4() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	// The filter step. It filters out odd numbers and reports the error.
	// you don't have to use the error handler, you can just return error to be discarded from
	// the pipeline
	step1 := pipelines.NewStepStandardWithErrorHandler("step1", 1, filterOdd, filterErrorHandler)
	//step1 := pipelines.NewStep("step1", 1, filterOdd) // filtering without printing error

	// The processing step
	step2 := pipelines.NewStepStandard("step2", 1, by2)

	// The result step
	resultStep := pipelines.NewStepResult("resultStep", 1, printFilterResult)

	// init pipeline
	pipe := pipelines.NewPipeline(10, step1, step2, resultStep)
	pipe.Init()

	// Running
	go pipe.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pipe.FeedOne(i)
	}

	// waiting for all tokens to be processed
	pipe.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipe.Terminate()

	// cancel the context can come before or after Terminate
	cancelCtx()

	fmt.Println("Example4 Done!!!")
}
