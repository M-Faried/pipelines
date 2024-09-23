# Motivation

The package is created to fulfill the need for having a pipeline with concurrent steps to facilitate the pipeline design pattern in Go programming language. Pipeline design pattern is generally useful when you have a complex process that you need to break down into multiple consecutive steps.

To put the package into perspective, I was working on a scientific application receiving a stream of data from another service and had to make heavy transformation on the data before saving in the data base.

I found myself having to handle complex transformation of data in addition to dealing with oncurrency and synchronization challenges at the same time. I also wanted to scale up some parts of the transformation equation by creating multiple go routines to serve, then feed their results to the following steps. I also wanted to filter some inputs during the execution based on a certain threshould.

### Benefits Of Using Pipelines

- Very simple to use, integrate, and extend.

- Built using generics to acomodate all data types.

- Elevates concurrency and handles step communication through channels.

- Ability to horizontally scale any step of the pipelines and have more replicas to handle heavy processing parts of your operation.

- Ability to have filter steps or make any step filter the items going through you pipeline.

### Installation

```shell
go get github.com/m-faried/pipelines
```

# Usage

A couple of examples are submitted in the examples folder. Only example 1 is demonstrated here. I left guiding comments on the rest of the examples.

#### Exampe Description:

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

func filterNegativeValues(i int64) (int64, error) {
    if i < 0 {
        return 0, fmt.Errorf("filtering: %d", i)
    }
    return i, nil
}

func printResult(i int64) error {
    fmt.Printf("Result: %d \n", i)
    return nil
}

func main() {

    // Creating steps
    plus5Step := pipelines.NewStep("plus5", 1, plus5)
    minus10Step := pipelines.NewStep("minus10", 1, minus10)
    filterStep := pipelines.NewStep("filter", 1, filterNegativeValues)
    printResultStep := pipelines.NewResultStep("printResult", 1, printResult)


    // Creating & init the pipeline
    pipe := pipelines.NewPipeline(10, printResultStep, plus5Step, minus10Step, filterStep)
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
}
```

# Explanation

### Defining Intermediate Steps

You first define all the intermediate steps of your pipeline. The creation of the steps requires 3 arguments at least

1. The label of the step
2. The number of replicas of the step
3. The process to be run in this step

```go
plus5Step := pipelines.NewStep("plus5", 1, plus5)
minus10Step := pipelines.NewStep("minus10", 1, minus10)
```

Another version of the NewStep constructor called NewStepWithErrorHandler is available to enable users submit an error handler for the step in case they happen. When is reported by the step process, both the step label and the error sent to the error handler and the item caused the problem is dropped.

```go
// Error handler signature
type ReportError func(string, error)

// To create a step with error handler
step := pipelines.NewStepWithErrorHandler("step1", 1, stepProcess, errorHandlerFunction)
```

### Defining Result Step

Result step is the final step in the pipeline and can return errors only. It also has the same two versions of the constructor with the same parameters order. It is the process where you save the results of the pipeline to the database, send it over the network, ....

```go
// Defining result step without error handler
resultStep := pipelines.NewResultStep("printResult", 1, printResult)

// With error handler
resultStep := pipelines.NewResultStepWithErrorHandler("printResult", 1, printResult, errorHandlerFunction)
```

### Pipeline Creation

The pipeline creation requires three arguments

1. The buffer size of the channels used to communication among the pipeline. Configure it depending on your needs.
2. The result step.
3. One or more intermediate steps.

```go
channelBufferSize := 10
pipe := pipelines.NewPipeline(channelBufferSize, resultStep, step1, step2, step3)
```

### Pipeline Running

The pipeline requires first a context to before you can run the pipeline. Define a suitable context for your case and then sendit to the Run function. The Run function doesn't need to run in a go subroutine as it is not blocking.

```go
ctx, cancelCtx := context.WithCancel(context.Background())
pipe.Run(ctx)
```

### Feeding Items Into Pipeline

There are 2 functions to feed data into the pipeline depending on your case.

```go
pipe.FeedOne(item)
pipe.FeedMany(items)
```

### Waiting Pipeline To Finish

To wait for the pipeline to be done with all the items fed into it, you can use the **blocking** call to the following function:

```go
pipe.WaitTillDone()
```

### Terminating Pipeline

When you want to terminate the pipeline use the following function. Note that it will terminate regardless the parent context is closed or not. And once it terminates, it can't be rerun again and you need to create another pipeline.

```go
pipe.Terminate()
```

# Notes

- The error handler function **should NOT** block the implementation for long or else it will block and delay the execution through the pipeline.
- The pipeline can operate on on type, but you can create a container structure to have a separate field for every step to set if you want to accumulate results of different types.
- To filter an item and discontinue its processing, you need to return error.
