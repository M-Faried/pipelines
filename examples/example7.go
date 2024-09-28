package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func evenNumberCriteria(i int64) bool {
	return i%2 == 0
}

func calculateEvenSum(i []int64) pip.StepBufferedProcessOutput[int64] {

	fmt.Println("calculateEvenSum Input: ", i)

	// The following is just for illustrating the default values of the output
	ouput := pip.StepBufferedProcessOutput[int64]{
		HasResult: false,
		Result:    0,
		Flush:     false,
	}

	// checking the threshold to start calculation
	if len(i) >= 10 {
		var sum int64
		for _, v := range i {
			sum += v
		}
		ouput.HasResult = true
		ouput.Result = sum
		ouput.Flush = true
	}

	return ouput
}

// Example7 demonstrates a pipeline which outputs the sum of every 10 even numbers recevied.
func Example7() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(&pip.StepFilterConfig[int64]{
		Label:        "filter",
		Replicas:     1,
		PassCriteria: evenNumberCriteria,
	})

	aggregator := builder.NewStep(&pip.StepBuffered[int64]{
		Label:      "aggregator",
		Replicas:   5,
		BufferSize: 10,
		// This means the buffer is going to retain all elements.
		PassThrough: false,
		// Notice that, since the aggregation interval is not set, you will need to have an
		// accurate threshould formula to avoid stalling the accumulator for long.
		// You can use either or both threshold and interval time based on your needs in other cases.
		InputTriggeredProcess: calculateEvenSum,
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

	for i := 1; i <= 20; i++ {
		pipeline.FeedOne(int64(i))
	}

	// waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 7 Done !!!")
}
