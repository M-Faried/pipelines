package examples

import (
	"context"
	"fmt"
	"time"

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

	// The steps used from example 1
	plus5Step := pipelines.NewStep("plus5", 1, plus5)
	minus10Step := pipelines.NewStep("minus10", 1, minus10)
	printResultStep := pipelines.NewResultStep("printResult", 1, printResult)
	pipe := pipelines.NewPipeline(10, printResultStep, plus5Step, minus10Step) // Notice 10 is the buffer size
	pipe.Init()

	// Running
	go pipe.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 500; i++ {
		pipe.FeedOne(int64(i))
	}
	// 	waiting for all tokens to be processed
	pipe.WaitTillDone()
	// cancel the context
	cancelCtx()
	// optional: wait till all channels are closed
	time.Sleep(1 * time.Second)
	fmt.Println("Example1 Done!!!")
}
