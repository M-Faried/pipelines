package examples

import (
	"context"
	"fmt"
	"time"

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

	// The filter step
	step1 := pipelines.NewStepWithErrorHandler("step1", 1, filterOdd, filterErrorHandler)

	// The processing step
	step2 := pipelines.NewStep("step2", 1, by2)

	// The result step
	resultStep := pipelines.NewResultStep("resultStep", 1, printFilterResult)

	// init pipeline
	pipe := pipelines.NewPipeline(10, resultStep, step1, step2)
	pipe.Init()

	// Running
	go pipe.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pipe.FeedOne(i)
	}

	// waiting for all tokens to be processed
	pipe.WaitTillDone()
	// cancel the context
	cancelCtx()
	// optional: wait till all channels are closed
	time.Sleep(1 * time.Second)
	fmt.Println("Example4 Done!!!")
}
