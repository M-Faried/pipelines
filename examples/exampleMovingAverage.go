package examples

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	pip "github.com/m-faried/pipelines"
)

func calculateAverage(buffer []int64) (int64, pip.BufferFlags) {

	fmt.Println("calculateEvenSum Input: ", buffer)

	var sum int64
	for _, v := range buffer {
		sum += v
	}

	avg := sum / int64(len(buffer))

	return avg, pip.BufferFlags{
		SendProcessOuput: true,
	}
}

func createMovingAveragePipeline() pip.IPipeline[int64] {
	builder := &pip.Builder[int64]{}

	filter := builder.NewStep(pip.StepFilterConfig[int64]{
		Label:            "filter",
		Replicas:         4,
		InputChannelSize: 8,
		PassCriteria: func(i int64) bool {
			return i%2 == 0
		},
	})

	buffer := builder.NewStep(pip.StepBufferConfig[int64]{
		Label:                 "buffer",
		Replicas:              1,
		BufferSize:            5,                // The moving average window size
		PassThrough:           false,            // This means that the buffer will not let the elements move to the following steps.
		InputTriggeredProcess: calculateAverage, // The process which is triggered with every input.
	})

	print := builder.NewStep(pip.StepTerminalConfig[int64]{
		Label:    "result",
		Replicas: 1,
		Process: func(token int64) {
			fmt.Println("Result:", token)
		},
	})

	pConfig := pip.PipelineConfig{
		DefaultStepInputChannelSize: 1,
	}
	pipeline := builder.NewPipeline(pConfig, filter, buffer, print)
	pipeline.Init()
	return pipeline
}

// ExampleMovingAverage demonstrates a pipeline which calculates the moving average of the latest 5 even numbers.
func ExampleMovingAverage() {

	// Set up a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancelCtx := context.WithCancel(context.Background())

	pipeline := createMovingAveragePipeline()
	pipeline.Run(ctx)

	// Goroutine to cancel the context when an interrupt signal is received
	go func() {
		<-sigChan
		fmt.Println("Interrupt signal received, cancelling context...")

		// cancel the context
		cancelCtx()

		// terminating the pipeline and clearning resources
		pipeline.Terminate()

		fmt.Println("Pipeline Terminated!!!")
	}()

	for i := 1; i <= 20; i++ {
		pipeline.FeedOne(int64(i))
	}

	fmt.Println("CTRL+C to exit!")

	<-ctx.Done()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Example Moving Average Done !!!")
}
