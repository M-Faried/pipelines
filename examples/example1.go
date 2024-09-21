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
func printResult(i int64) {
	fmt.Printf("Result: %d \n", i)
}

////////////////////////////

func Example1() {
	ctx, cancel := context.WithCancel(context.Background())

	plus5Step := pipelines.NewStep("plus5", 1, plus5)
	minus10Step := pipelines.NewStep("minus10", 1, minus10)
	printResultStep := pipelines.NewResultStep("printResult", 1, printResult)
	pip := pipelines.NewPipelineWithResultHandler(10, printResultStep, plus5Step, minus10Step)
	pip.Init()

	// Running
	go logProcess(pip)
	go pip.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 10; i++ {
		time.Sleep(1 * time.Second)
		pip.FeedOne(int64(i))
	}

	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("Example1 Done!!!")
}
