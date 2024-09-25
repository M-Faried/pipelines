# Motivation

The package is created to fulfil the need for having a pipeline with concurrent steps to facilitate the pipeline design pattern in Go programming language. Pipeline design pattern is generally useful when you have a complex process that you need to break down into multiple consecutive steps.

To put the package into perspective, I was working on a scientific application receiving a data stream from another service. Demanding and complex transformation was required on the data before saving it to the database.

I found myself having to handle complex transformation of data in addition to dealing with concurrency and synchronization challenges. I also needed to scale up some parts of the transformation equation and make their execution concurrent in addition to limiting the writing rate to the database. And, of course, different variations of the transformation were required.

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

A couple of examples are submitted in the examples folder in addition to the following example.

#### Exampe Description:

1. Add 5 to the number

2. Subtract 10 from the number

3. Filter negative values

4. Print result

```Go

import (
	"context"
	"fmt"

	pip "github.com/m-faried/pipelines"
)

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

    builder := &pip.Builder[int64]{}

    // Creating steps
    plus5Step := builder.NewStep(&pip.StepConfig[int64]{
        Label:    "plus5",
        Replicas: 1,
        Process:  plus5,
    })

    minus10Step := builder.NewStep(&pip.StepConfig[int64]{
        Label:    "minus10",
        Replicas: 1,
        Process:  minus10,
    })

    filterStep := builder.NewStep(&pip.StepConfig[int64]{
        Label: "filter",
        Replicas: 1,
        Process: filterNegativeValues,
    })

    printResultStep := builder.NewStep(&pip.StepResultConfig[int64]{
        Label:    "print",
        Replicas: 1,
        Process:  printResult,
    })


    // Creating & init the pipeline
    channelsBufferSize := 10
    pipeline := builder.NewPipeline(channelsBufferSize, plus5Step, minus10Step, filterStep, printResultStep)
    pipeline.Init()


    // Running
    ctx := context.Background()
    pipeline.Run(ctx)


    // Feeding inputs
    for i := 0; i <= 50; i++ {
        pipeline.FeedOne(int64(i))
    }


    // Waiting for all tokens to be processed
    pipeline.WaitTillDone()

    // Terminating the pipeline and clearning resources
    pipeline.Terminate()
}
```

# Explanation

### Defining Intermediate Steps

#### You have 3 types of steps:

1. **Single Input Single Output Step** which is created using **NewStep**

2. **Single Input Multiple Output Step** which is created using **NewStepFragmenter** (example 5)

3. **Single Input No Output Step** which is created using **NewStepResult**

You first define all the intermediate steps of your pipeline. The creation of the steps **requires the core process and will panic if not submitted**, but you can add also optionally:

- Label (empty string by default and needed for error reporting)

- Replicas count (1 by default)

- Error handler (nil by default and will not be called if not set)

```go
// The process type expected by the step
// type StepProcess[I any] func(I) (I, error)

plus5Step := builder.NewStep(&pip.StepConfig[int64]{
    Process:  plus5,
})

minus10Step := builder.NewStep(&pip.StepConfig[int64]{
    Label:    "minus10",
    Replicas: 3,
    Process:  minus10,
})
```

### Error Handler

You can add the error handler to the configuration of the step in case you need to handle error. When is reported by the step process, both the step label and the error sent to the error handler and the item caused the problem is dropped.

This applies to all configurations of steps.

```go
// The handler type expected as error handler.
// type ErrorHandler func(label string, err error)

// To create a step with error handler
plus5Step := builder.NewStep(&pip.StepConfig[int64]{
    Label:    "plus5",
    Process:  plus5,
    ErrorHandler: func(label string, err error ){
        // the body of the error handler
    },
})
```

### Defining Result Step

Result step is the final step in the pipeline and can return errors only. It also has the same two versions of the constructor with the same parameters order. It is the process where you save the results of the pipeline to the database, send it over the network, ....

```go
// The result process type expected by the result step.
// type StepResultProcess[I any] func(I) error

// result step
printResultStep := builder.NewStep(&pip.StepResultConfig[int64]{
    Label:    "print",
    Replicas: 1,
    Process:  printResult,
})
```

### Defining Fragmenter Step (Example 5)

Fragmenter step allows users to break down a token into multiple tokens and fed to the following steps of the pipelines. This is useful when you are having a large chunk of data and want to breake down into a smaller problem and aggregate results later into the result step. You can reap its benifits of course when you increase the number of the replcias of subsequent steps.

```go
// The fragmenter process type expected by the fragmenter step.
// type StepFragmenterProcess[I any] func(I) ([]I, error)

builder := &pip.Builder[string]{}
// fragmenter step
splitter := builder.NewStep(&pip.StepFragmenterConfig[string]{
    Label:    "fragmenter",
    Replicas: 1,
    Process:  splitter,
})
```

### Pipeline Creation

The pipeline creation requires 2 arguments

1. The buffer size of the channels used to communication among the pipeline. Configure it depending on your needs.

2. Steps in the order of execution

```go
channelBufferSize := 10
pipeline := builder.NewPipeline(channelBufferSize, step1, step2, step3, resultStep)
```

### Pipeline Running

The pipeline requires first a context to before you can run the pipeline. Define a suitable context for your case and then sendit to the Run function. The Run function doesn't need to run in a go subroutine as it is not blocking.

```go
ctx := context.Background() // any type of context can be used here
pipeline.Run(ctx)
```

### Feeding Items Into Pipeline

There are 2 functions to feed data into the pipeline depending on your case.

```go
pipeline.FeedOne(item)
pipeline.FeedMany(items)
```

### Waiting Pipeline To Finish

To wait for the pipeline to be done with all the items fed into it, you can use the **blocking** call to the following function:

```go
pipeline.WaitTillDone()
```

### Terminating Pipeline

When you want to terminate the pipeline use the following function. Note that it will terminate regardless the parent context is closed or not. And once it terminates, it can't be rerun again and you need to create another pipeline.

```go
pipeline.Terminate()
```

# Notes

- The error handler function **should NOT** block the implementation for long or else it will block and delay the execution through the pipeline. The error can be transferred to another function or process to be handled if it is going to take long.

- The pipeline can operate on on type, but you can create a container structure to have a separate field for every step to set if you want to accumulate results of different types.

- To filter an item and discontinue its processing, you need to return error.
