package examples

import (
	"context"
	"fmt"
	"time"

	pip "github.com/m-faried/pipelines"
)

func oddNumberCriteria(i int64) bool {
	return i%2 != 0
}

func aggOddSumProcess(i []int64) (int64, error) {
	var sum int64
	for _, v := range i {
		sum += v
	}
	return sum, nil
}

// Example7 demonstrates a pipeline with an aggregator step.
// The aggregator step buffers the odd number and calculates their summation.
func Example7() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(&pip.StepFilterConfig[int64]{
		Label:        "filter",
		Replicas:     1,
		PassCriteria: oddNumberCriteria,
	})

	aggregator := builder.NewStep(&pip.StepAggregatorConfig[int64]{
		Label:    "aggregator",
		Replicas: 3,
		Process:  aggOddSumProcess,
		// ThresholdCriteria: func(items []int64) bool { return len(items) >= 10 },
		// Notice that, since the threshold is not set, you will need to have an
		// accurate interval time for inputs to avoid stalling the aggregator for long.
		// You can use either or both threshold and interval time based on your needs.
		AggregationInterval: 150 * time.Millisecond, //This means the aggregator calculates the result every 150ms

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

	fmt.Println("Example 7 Done !!!")
}
