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

func periodicCalculateSum(buffer []int64) (int64, pip.BufferFlags) {
	fmt.Println("calculateOddSum Input: ", buffer)

	var sum int64
	for _, v := range buffer {
		sum += v
	}

	return sum, pip.BufferFlags{
		SendProcessOuput: true,
		FlushBuffer:      false,
	}
}

// Example 6 calclulates the sum of the most recently received 5 odd numbers every 100ms.
func Example6() {

	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(pip.StepFilterConfig[int64]{
		Label:        "filter",
		Replicas:     1,
		PassCriteria: oddNumberCriteria,
	})

	sumCalculator := builder.NewStep(pip.StepBufferConfig[int64]{
		Label:      "aggregator",
		Replicas:   2,
		BufferSize: 5,
		// This means the buffer is going to retain all elements.
		PassThrough: false,
		// Notice that, since the InputTriggeredProcess is not set, you will need to have an
		// accurate interval time for inputs to avoid stalling pipeline for long.
		// You can use either or both threshold and interval time based on your needs in other cases.
		TimeTriggeredProcess:         periodicCalculateSum,
		TimeTriggeredProcessInterval: 500 * time.Millisecond, //This means the buffer calculates the result from the buffer every 500ms

	})

	result := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "result",
		Replicas: 1,
		Process:  printResult,
	})

	pConfig := pip.PipelineConfig{
		DefaultStepInputChannelSize: 10,
	}
	pipeline := builder.NewPipeline(pConfig, filter, sumCalculator, result)
	pipeline.Init()

	ctx := context.Background()
	pipeline.Run(ctx)

	for i := 1; i <= 20; i++ {
		pipeline.FeedOne(int64(i))
		time.Sleep(1 * time.Second)
	}

	// since the values of the buffere are not flushed, using WaitTillDone will keep your pipeline running forever
	// pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 6 Done !!!")
}
