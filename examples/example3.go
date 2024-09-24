package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/m-faried/pipelines"
)

func by10(i int64) (int64, error) {
	return i * 10, nil
}

func by100(i int64) (int64, error) {
	time.Sleep(2 * time.Second)
	return i * 100, nil
}

func printBy10Result(i int64) error {
	fmt.Printf("Result: %d \n", i)
	return nil
}

// Example3 demonstrates how to utilize replicas feature when you have a heavy process.
func Example3() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	step1 := pipelines.NewStep("step1", 1, by10)
	step2 := pipelines.NewStep("step2", 10, by100) // Heavy process so we need 10 replicas
	resultStep := pipelines.NewStepResult("resultStep", 1, printBy10Result)
	pipe := pipelines.NewPipeline(10, step1, step2, resultStep)
	pipe.Init()

	// Running
	pipe.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pipe.FeedOne(i)
	}

	// Notice that we need only 3 seconds to process all inputs although by100 process takes 2 seconds
	// and we process 10 items.
	time.Sleep(3 * time.Second)

	// we can replace the previous step with the following, but it was omitted to show the power of replicas
	// pipe.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipe.Terminate()

	// cancel the context can come before or after Terminate
	cancelCtx()

	fmt.Println("Example3 Done!!!")
}
