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

More examples are submitted in the examples folder in addition to the following simple one.

#### Simple Example:

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

func isPositiveValue(i int64) bool {
    return i >= 0
}

func printResult(i int64) {
    fmt.Printf("Result: %d \n", i)
    return nil
}

func main() {

    builder := &pip.Builder[int64]{}

    // Creating steps
    plus5Step := builder.NewStep(pip.StepBasicConfig[int64]{
        Label:    "plus5",
        Replicas: 1,
        Process:  plus5,
    })

    minus10Step := builder.NewStep(pip.StepBasicConfig[int64]{
        Label:    "minus10",
        Replicas: 1,
        Process:  minus10,
    })

    filterStep := builder.NewStep(pip.StepFilterConfig[int64]{
        Label: "filter",
        Replicas: 1,
        PassCriteria: isPositiveValue,
    })

    printResultStep := builder.NewStep(pip.StepResultConfig[int64]{
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

### Types Of Steps:

1. **Basic Step:** Carries on a transformation on a single token in the pipeline.

2. **Filter Step:** Filters some undesired input to the pipeline.

3. **Result Step:** The final step in the pipeline where the results of previous steps are presented or saved.

4. **Fragmenter Step:** Breaks down any token into multiple tokens and feeds them to the next steps in the pipeline.

5. **Buffered Step:** Retains multiple elements in the pipeline to run a calculation over periodically or based on input.

Based on the type of the step your create, different configurations are required to be submitted by the user. One of these configurations is the pipelines which allows to scale up any step of any type just by setting the number of replicas of each step in the configuration.

## Builder

You need to create an instance first from the builder to use it to create any part of the pipeline.

```go
builder := &pip.Builder[<TypeOfPipelineAndSteps>]{}
```

## Basic Step

Carries on a transformation on a single token in the pipeline and pushes it forward to the next steps.

- Label (empty string by default and needed for error reporting and future use)

- Replicas count (The number of replicas to be run for this step)

- The transformation process to be run on every token in the pipeline

```go
// The process type expected by the step
type StepBasicProcess[I any] func(I) (I, error) // Don't redefine

step := builder.NewStep(pip.StepBasicConfig[int64]{
    Replicas: 3,
    Label:    "minus10",
    Process:  func(token int64) (int64, error) {
        return token - 10, nil
    }, //StepProcess
})
```

### Error Handler (Available For Basic Step Only)

You can add the error handler to the configuration of the basic step in case you need to handle errors. When is reported by the step process, both the step label and the error sent to the error handler and the item caused the problem is dropped.

```go
// StepBasicErrorHandler is the definition of error reporting handler which may or may not be set by the user during creation of the step.
// The first parameter is the label of the step where the error occurred and the second parameter is the error itself.
type StepBasicErrorHandler func(string, error) // Don't redefine

// To create a step with error handler
step := builder.NewStep(pip.StepBasicConfig[int64]{
    Replicas:       4,
    Label:          "plus5",
    Process:        plus5,
    ErrorHandler:   func(label string, err error ){
        // the body of the error handler here
    },
})
```

## Filter Step

Used to get rid of any undesired tokens from the pipelines

```go
// StepFilterPassCriteria is function that determines if the data should be passed or not.
type StepFilterPassCriteria[I any] func(I) bool // Don't redefine

step := builder.NewStep(pip.StepFilterConfig[int64]{
    Replicas:     1,
    Label:        "filterEven",
    PassCriteria: func(token int64) bool {
        return token%2 == 0
    },
})
```

## Result Step

Result step is the final step in the pipeline and can return errors only. It also has the same two versions of the constructor with the same parameters order. It is the process where you save the results of the pipeline to the database, send it over the network, ....

```go
// The result process type expected by the result step.
type StepResultProcess[I any] func(I) // Don't redefine

// result step
resultStep := builder.NewStep(pip.StepResultConfig[int64]{
    Replicas: 1,
    Label:    "print",
    Process:  func(token int64) {
        fmt.Println("Result:", token)
    },
})
```

## Fragmenter Step (Example 5)

Fragmenter step allows users to break down a token into multiple tokens and fed to the following steps of the pipelines. This is useful when you are having a large chunk of data and want to breake down into a smaller problem and aggregate results later into the result step. You can reap its benifits of course when you increase the number of the replcias of subsequent steps.

```go
// The fragmenter process type expected by the fragmenter step.
type StepFragmenterProcess[I any] func(I) []I // Don't redefine

// fragmenter step
splitter := builder.NewStep(pip.StepFragmenterConfig[string]{
    Label:    "fragmenter",
    Replicas: 1,
    Process:  func(token string) []string {
        // Split the Value field by comma
	    splitTokens := strings.Split(token.Value, ",")
        return splitTokens
    },
})
```

## Buffered Step (Examples 6 & 7)

Buffered step is a step that doesn't apply an operation on a single token, but buffers multiple items and act on them accordingly. The **collected tokens span** can be of 2 types:

1. Received Inputs Count Span
2. Periodic Time Span (The received token over a period of time)

The buffered step supports either or both of the 2 kinds of span. It also supports the pass through feature which allows users of the package to control whether the buffered pipeline tokens are sent to the following steps after being buffered or not.

You have also the option to flush the buffer or not after running your calculation on the buffer.

### Time Triggered Buffer Step (Example 6)

In the time triggered step, every input is added to the buffer, but the process **TimeTriggeredProcess** is called with the current buffer value every **TimeTriggeredProcessInterval** set in the step configuration.

In the following example, the moving sum is calculated with window of size 10.

```go

// StepBufferedProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferedProcess[I any] func([]I) StepBufferedProcessOutput[I] // Don't redefine

func periodicCalculateSum(buffer []int64) pip.StepBufferedProcessOutput[int64] {
	var sum int64
	for _, v := range buffer {
		sum += v
	}
	return pip.StepBufferedProcessOutput[int64]{
		HasResult:   true,
		Result:      sum,
		FlushBuffer: false,
	}
}

bufferStep := builder.NewStep(pip.StepBufferedConfig[int64]{
    Label:      "periodic buffer",
    Replicas:   2,
    BufferSize: 5,
    PassThrough:                    false,// This means the buffer is going to retain all elements.
    TimeTriggeredProcess:           periodicCalculateSum, // The process which is called periodically with the current buffer value.
    TimeTriggeredProcessInterval:   100 * time.Millisecond, //This means the buffer calculates the result from the buffer every 500ms

})
```

### Input Triggered Buffer Step (Example 7)

For the input triggered buffer, every input received is added to the buffer and then **InputTriggeredProcess** is called with the most recent buffer value.

In the following example, process waits till the buffer is full and calculates the sum with flushing the elements already added to the buffer.

```go
// StepBufferedProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferedProcess[I any] func([]I) StepBufferedProcessOutput[I] // Don't redefine

func calculateSumOnBufferCountThreshold(buffer []int64) pip.StepBufferedProcessOutput[int64] {

    // skipping the calcluation if not enough data was buffered.
    if len(buffer) < 10 {
        return pip.StepBufferedProcessOutput[int64]{
            HasResult:   false,
            Result:      0,
            FlushBuffer: false,
	    }
    }

    var sum int64
    for _, v := range buffer {
        sum += v
    }
    return pip.StepBufferedProcessOutput[int64]{
        HasResult:   true,
        Result:      sum,
        FlushBuffer: true, // This instructs the buffered step to flush the data it retains and start accumulating fresh data.
    }
}

bufferStep := builder.NewStep(pip.StepBufferedConfig[int64]{
    Label:      "buffer",
    Replicas:   5,
    BufferSize: 10,
    PassThrough:            false,// This means the buffer is going to retain all elements.
    InputTriggeredProcess:  calculateSumOnBufferCountThreshold,
})
```

### Important Notes

- The buffer size remains fixed, and when it is full the **oldest data is overwritten**. You will find the oldest data at index 0 and the most recent at the last element of the array.

- Although you can set replicas, but all operating on the concurrent safe buffer.

- The InputTriggered and the TimeTriggered processes are implemented in a concurrent safe context so the buffer will not change during their execution.

- When you set the **PassThrough** to true and the input triggered process returns a valid result. The received input will be sent first to the following steps then the new result.

- Again, you can set both time triggered and input triggered processes for the buffer step and they will be both be executed by their triggeres.

## Pipeline

Although the pipeline can operate on one type, but you can create a container structure to have a separate field for every step to set if you want to accumulate results of different types.

The pipeline creation requires 2 arguments

1. The buffer size of the channels used to communication among the pipeline. Configure it depending on your needs.

2. Steps in the order of execution.

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

To wait for the pipeline to be done with all the items fed into it, you can use the **blocking** call to wait before resuming execution of your program:

```go
pipeline.WaitTillDone()
```

It is an optional step to use and you can call Terminate() directly without waiting, but all the tokens in the pipeline will be discarded.

When using buffered channels don't use **pipeline.WaitTillDone()** unless you have a finite number of inputs and you flush the data in the buffer regularly. Otherwise wait till done will stall your application and may result a deadlock..

### Terminating Pipeline

When you want to terminate the pipeline use the following function. Note that it will terminate regardless the parent context is closed or not. And once it terminates, it can't be rerun again and you need to create another pipeline.

```go
pipeline.Terminate()
```

# Real Life Example (Sensor Data)

In the examples folder you will find a full and complex example on how to use pipelines to process data streams efficiently. The example is the code to solve the below problem statement.

### Problem Statement:

- We have a sensor that sends data every 50ms in the structure SensorData. The data which should be saved in the database is the average of valid readings received from the sensor over 1 second span and the data should be approximated.

- If no data is received the average should still be logged with the most recent average value. And if the average is never calculated before, during the init phase, averages should still be logged with zero in the db.

- During the initialization of the sensor, the sensor goes through a calibration phase during which the sensor sends invalid temprature and humidity values which should be filtered. The invalid data value of the temprature less than -100 or the humidity is less than 0.

- The error data in the sensor reading should be monitored and calculated according to the following:

  - If the average error of the latest 5 temprature readings exceeds value of "10", the sensor should be recalibrated.

  - If the average error of the latest 10 humidity readings exceeds value of "15", the sensor should be recalibrated.
