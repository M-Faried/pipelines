package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/m-faried/pipelines"
)

func Example4() {
	ctx, cancel := context.WithCancel(context.Background())

	// The steps used from example 1
	plus5Step := pipelines.NewStep("plus5", 1, plus5)
	minus10Step := pipelines.NewStep("minus10", 1, minus10)
	printResultStep := pipelines.NewResultStep("printResult", 1, printResult)
	pip := pipelines.NewPipeline(100, printResultStep, plus5Step, minus10Step)
	pip.Init()

	// Running
	go pip.Run(ctx)

	// Feeding inputs
	for i := 0; i <= 500; i++ {
		pip.FeedOne(int64(i))
	}

	// you can wait till the pipeline is done like this
	for {
		count := pip.TokensCount()
		if count == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	cancel()
	time.Sleep(1 * time.Second)
	fmt.Println("Example1 Done!!!")
}
