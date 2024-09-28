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

func calculateOddSum(i []int64) pip.TimeTriggeredProcessOutput[int64] {

	fmt.Println("calculateOddSum Input: ", i)

	// The following is just for illustrating the default values of the output
	result := pip.TimeTriggeredProcessOutput[int64]{
		HasResult: false,
		Result:    0,
		Flush:     false,
	}

	var sum int64
	for _, v := range i {
		sum += v
	}

	result.HasResult = true
	result.Result = sum
	result.Flush = false //without flush since it will be time triggered process.

	return result
}

// Example 6 calclulates the sum of the most recently received 5 odd numbers every 100ms.
func Example6() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(&pip.StepFilterConfig[int64]{
		Label:        "filter",
		Replicas:     1,
		PassCriteria: oddNumberCriteria,
	})

	aggregator := builder.NewStep(&pip.StepBuffered[int64]{
		Label:      "aggregator",
		Replicas:   2,
		BufferSize: 5,
		// Notice that, since the InputTriggeredProcess is not set, you will need to have an
		// accurate interval time for inputs to avoid stalling pipeline for long.
		// You can use either or both threshold and interval time based on your needs in other cases.
		TimeTriggeredProcess:         calculateOddSum,
		TimeTriggeredProcessInterval: 100 * time.Millisecond, //This means the buffer calculates the result from the buffer every 500ms

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
		time.Sleep(500 * time.Millisecond)
	}

	// since the values of the buffere are not flushed, using WaitTillDone will keep your pipeline running forever
	// pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 6 Done !!!")
}
