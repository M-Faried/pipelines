package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

func evenNumberCriteria(i int64) bool {
	return i%2 == 0
}

func calculateSumOnBufferCountThreshold(buffer []int64) (int64, pip.BufferFlags) {

	fmt.Println("calculateEvenSum Input: ", buffer)

	// checking the threshold to start calculation
	if len(buffer) >= 10 {
		var sum int64
		for _, v := range buffer {
			sum += v
		}
		return sum, pip.BufferFlags{
			SendProcessOuput: true,
			FlushBuffer:      true, // This means the buffer is going to be flushed after the calculation.
		}
	}

	return 0, pip.BufferFlags{
		SendProcessOuput: false,
		FlushBuffer:      false,
	}
}

// Example7 demonstrates a pipeline which outputs the sum of every 10 even numbers recevied.
func Example7() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(pip.StepFilterConfig[int64]{
		Label:        "filter",
		Replicas:     1,
		PassCriteria: evenNumberCriteria,
	})

	buffer := builder.NewStep(pip.StepBufferConfig[int64]{
		Label:      "aggregator",
		Replicas:   5,
		BufferSize: 10,
		// This means the buffer is going to retain all elements.
		PassThrough: false,
		// Notice that, since the aggregation interval is not set, you will need to have an
		// accurate threshould formula to avoid stalling the accumulator for long.
		// You can use either or both threshold and interval time based on your needs in other cases.
		InputTriggeredProcess: calculateSumOnBufferCountThreshold,
	})

	result := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "result",
		Replicas: 1,
		Process:  printResult,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepChannelSize: 10,
		TrackTokensCount:       true,
	}
	pipeline := builder.NewPipeline(pConfig, filter, buffer, result)
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
