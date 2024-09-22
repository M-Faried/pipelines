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

func printBy10(i int64) error {
	fmt.Printf("Result: %d \n", i)
	return nil
}

func Example3() {
	ctx, cancel := context.WithCancel(context.Background())

	step1 := pipelines.NewStep("step1", 1, by10)
	step2 := pipelines.NewStep("step2", 10, by100)
	resultStep := pipelines.NewResultStep("resultStep", 1, printBy10)
	pip := pipelines.NewPipelineWithResultHandler(100, resultStep, step1, step2)
	pip.Init()

	// Running
	go pip.Run(ctx)

	// Feeding inputs
	for i := int64(0); i < 10; i++ {
		pip.FeedOne(i)
	}

	time.Sleep(3 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("Example3 Done!!!")
}
