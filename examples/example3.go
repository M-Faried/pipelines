package examples

import (
	"context"
	"fmt"
	"time"

	pip "github.com/m-faried/pipelines"
)

func by10(i int64) int64 {
	return i * 10
}

func by100(i int64) int64 {
	time.Sleep(2 * time.Second)
	return i * 100
}

func printBy10Result(i int64) {
	fmt.Printf("Result: %d \n", i)
}

// Example3 demonstrates how to utilize replicas feature when you have a heavy process.
func Example3() {

	builder := &pip.Builder[int64]{}

	step1 := builder.NewStep(pip.StepBasicConfig[int64]{
		Label:    "step1",
		Replicas: 1,
		Process:  by10,
	})
	step2 := builder.NewStep(pip.StepBasicConfig[int64]{
		Label:    "step2",
		Replicas: 10, // Heavy process so we need 10 replicas
		Process:  by100,
	})
	stepResult := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "result",
		Replicas: 1,
		Process:  printBy10Result,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepInputChannelSize: 10,
	}
	pipeline := builder.NewPipeline(pConfig, step1, step2, stepResult)
	pipeline.Init()

	// Running
	ctx := context.Background()
	pipeline.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pipeline.FeedOne(i)
	}

	// Notice that we need only 3 seconds to process all inputs although by100 process takes 2 seconds
	// and we process 10 items.
	time.Sleep(3 * time.Second)

	// we can replace the previous step with the following, but it was omitted to show the power of replicas
	// pipe.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 3 Done!!!")
}
