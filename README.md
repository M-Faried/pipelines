# Motivation

The package is created to fulfill the need for having a simple pipeline with a simple interface to facilitate the pipeline design pattern in Go programming language.

Pipeline design pattern is generally useful when you have a complex process that you need to break down into multiple consecutive steps.

# Benefits Of Pipelines Package

- Very simple to use, integrate, and extend.
- Built using generic to acomodate all data types.
- Elevates concurrency and handles step communication for you through channels.
- Ability to horizontally scale any step of the pipelines and have more replicas to handle heavy processing parts of your operation.
- Ability to have filter steps or make any step filter the items going through you pipeline.

# Installation

```shellscript
go get github.com/m-faried/pipelines
```

# Usage

A couple of examples are submitted in the examples folder. Only example 1 is demonstrated here. I left guiding comments on the rest of the examples.

# Example

#### Process Description:

1. Add 5 to the number
2. Subtract 10 from the number
3. Print result

```Go
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
```

#### Pipeline Creation

```go

plus5Step := pipelines.NewStep("plus5", 1, plus5)

minus10Step := pipelines.NewStep("minus10", 1, minus10)

printResultStep := pipelines.NewResultStep("printResult", 1, printResult)

pipe := pipelines.NewPipeline(10, printResultStep, plus5Step, minus10Step)
pipe.Init()


// Running
ctx, cancelCtx := context.WithCancel(context.Background())
pipe.Run(ctx)


// Feeding inputs
for i := 0; i <= 50; i++ {
	pipe.FeedOne(int64(i))
}


// Waiting for all tokens to be processed
pipe.WaitTillDone()

// Cancel the context can come before or after Terminate
cancelCtx()

// Terminating the pipeline and clearning resources
pipe.Terminate()

```

# Notes

- The pipeline can operate on on type, but you can create a container structure to have a separate field for every step to set if you want to accumulate results of different types.
- To filter an item and discontinue its processing, you need to return error.
- Error handling is left to the user of the package, and you can have a different handler for each step using **NewStepWithErrorHandler** instead of NewStep.
