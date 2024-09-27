package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func evenNumberCriteria(i int64) bool {
	return i%2 == 0
}

func aggEvenThreshould(i []int64) bool {
	return len(i) >= 10
}

func aggEvenSumProcess(i []int64) (int64, error) {
	var sum int64
	for _, v := range i {
		sum += v
	}
	return sum, nil
}

// Example6 demonstrates a pipeline with an aggregator step.
// The aggregator step accumulates the even number and calculates their summation.
func Example6() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(&pip.StepFilterConfig[int64]{
		Label:    "filter",
		Replicas: 1,
		Process:  evenNumberCriteria,
	})

	aggregator := builder.NewStep(&pip.StepAggregatorConfig[int64]{
		Label:             "aggregator",
		Replicas:          3,
		Process:           aggEvenSumProcess,
		ThresholdCriteria: aggEvenThreshould,
		// AggregationInterval: 0,
		// Notice that, since the wait time is not set, you will need to have an
		// accurate threshould formula to avoid stalling the accumulator for long.
		// You can use either or both threshold and interval time based on your needs.
	})

	result := builder.NewStep(&pip.StepResultConfig[int64]{
		Label:    "result",
		Replicas: 1,
		Process:  printResult,
	})

	pipeline := builder.NewPipeline(10, filter, aggregator, result)
	pipeline.Init()

	ctx := context.Background()
	pipeline.Run(ctx)

	for i := 0; i < 20; i++ {
		pipeline.FeedOne(int64(i))
	}

	// waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 6 Done !!!")
}
