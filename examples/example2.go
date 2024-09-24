package examples

import (
	"context"
	"fmt"

	"github.com/m-faried/pipelines"
)

type Token struct {
	values    []string
	currValue string
	count     int
}

func processToken(t *Token) (*Token, error) {
	t.currValue = fmt.Sprintf("*%s*", t.currValue)
	t.values = append(t.values, t.currValue)
	t.count++
	return t, nil
}

func printToken(t *Token) error {
	fmt.Println("Result:", t.currValue)
	return nil
}

// Example2 demonstrates how to use pipelines with custom struct as input and output.
// It also demonstrates how to create identical steps in a pipeline.
func Example2() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	step1 := pipelines.NewStep("step1", 1, processToken)
	step2 := pipelines.NewStep("step2", 1, processToken)
	step3 := pipelines.NewStep("step3", 1, processToken)
	resultStep := pipelines.NewStepResult("outStep", 1, printToken)
	pipe := pipelines.NewPipeline(10, step1, step2, step3, resultStep)
	pipe.Init()

	// Running
	pipe.Run(ctx)

	// Feeding inputs
	pipe.FeedOne(&Token{values: []string{}, currValue: "Hello"})
	pipe.FeedOne(&Token{values: []string{}, currValue: "World"})
	pipe.FeedMany([]*Token{
		{values: []string{}, currValue: "Welcome"},
		{values: []string{}, currValue: "All"},
	})

	// waiting for all tokens to be processed
	pipe.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipe.Terminate()

	// cancel the context can come before or after Terminate
	cancelCtx()

	fmt.Println("Example2 Done!!!")
}
