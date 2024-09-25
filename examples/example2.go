package examples

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
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

	builder := &pip.Builder[*Token]{}

	step1 := builder.NewStep(&pip.StepConfig[*Token]{
		Label:    "step1",
		Replicas: 1,
		Process:  processToken,
	})
	step2 := builder.NewStep(&pip.StepConfig[*Token]{
		Label:    "step2",
		Replicas: 1,
		Process:  processToken,
	})
	step3 := builder.NewStep(&pip.StepConfig[*Token]{
		Label:    "step3",
		Replicas: 1,
		Process:  processToken,
	})
	resultStep := builder.NewStep(&pip.StepResultConfig[*Token]{
		Label:    "result",
		Replicas: 1,
		Process:  printToken,
	})

	pipeline := builder.NewPipeline(10, step1, step2, step3, resultStep)
	pipeline.Init()

	// Running
	ctx := context.Background()
	pipeline.Run(ctx)

	// Feeding inputs
	pipeline.FeedOne(&Token{values: []string{}, currValue: "Hello"})
	pipeline.FeedOne(&Token{values: []string{}, currValue: "World"})
	pipeline.FeedMany([]*Token{
		{values: []string{}, currValue: "Welcome"},
		{values: []string{}, currValue: "All"},
	})

	// waiting for all tokens to be processed
	pipeline.WaitTillDone()

	// terminating the pipeline and clearning resources
	pipeline.Terminate()

	fmt.Println("Example 2 Done!!!")
}
