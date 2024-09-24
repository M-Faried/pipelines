package examples

import (
	"context"
	"fmt"

	"github.com/m-faried/pipelines"
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
	ctx, cancelCtx := context.WithCancel(context.Background())

	plus5Step := pipelines.NewStep("plus5", 1, plus5)
	minus10Step := pipelines.NewStep("minus10", 1, minus10)
	printResultStep := pipelines.NewStepResult("printResult", 1, printResult)
	pipe := pipelines.NewPipeline[int64](10, plus5Step, minus10Step, printResultStep) // Notice 10 is the buffer size
	pipe.Init()

	// Running
	pipe.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 500; i++ {
		pipe.FeedOne(int64(i))
	}

	// 	waiting for all tokens to be processed
	pipe.WaitTillDone()

	// cancel the context can come before or after Terminate
	cancelCtx()

	// terminating the pipeline and clearning resources
	pipe.Terminate()

	fmt.Println("Example1 Done!!!")
}
