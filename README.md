# Motivation

<p align="center">
  <img src="logo.png" alt="Logo" />
</p>

The package is created to fulfil the need for having a pipeline with concurrent steps to facilitate the pipeline design pattern in Go programming language. Pipeline design pattern is generally useful when you have a complex process that you need to break down into multiple consecutive steps.

To put the package into perspective, I was working on a scientific application receiving a data stream from another service. A complex and demanding data transformation was required before saving it to the database.

I found myself having to handle complex complex data transformations in addition to dealing with concurrency and synchronization challenges.

Scaling up the whole execution of the process and having multiple concurrent workers was an option, but some steps in the execution needed to aggregate large chunks of data that might be fragmented across workers and to synchronize them was a nightmarish experience.

So I built this package to separate and hide scaling, concurrency, and synchronization concerns from the complex transformation process.

### Benefits Of Using Pipelines Package

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

func plus5(i int64) int64 {
    return i + 5
}

func minus10(i int64) int64 {
    return i - 10
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

    printResultStep := builder.NewStep(pip.StepTerminalConfig[int64]{
        Label:    "print",
        Replicas: 1,
        Process:  printResult,
    })

    // Creating & init the pipeline
    config := pip.PipelineConfig{
        DefaultStepInputChannelSize: 10,
        TrackTokensCount:       true,
    }
    pipeline := builder.NewPipeline(config, plus5Step, minus10Step, filterStep, printResultStep)
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

3. **Terminal Step:** The final step in the pipeline where the results of pipeline steps are presented, saved, or sent over network.

4. **Fragmenter Step:** Breaks down any token into multiple tokens and feeds them to the next steps in the pipeline.

5. **Buffer Step:** Retains multiple elements in the pipeline to run a calculation over periodically or based on input.

Based on the type of the step your create, different configurations are required to be submitted by the user.

### All Steps Basic Configuration:

Each step has basic configuration in addition to other configuration based on the type of the required step. These basic configuration are:

1. **Label:** Used by the user to identify the step and give it context specific name and set to empty string by default.

2. **Replicas:** Used to scale up any step of any type. It's optional and its default value is set to 1. It specifies the number of go routines to run steps in.

3. **InputChannelSize:** Specifies the buffer size of the input channel to the step. It is optional, and if it is not set, the input channel buffer size will be set to the default channel size set in the pipeline creation.

## Builder

You need to create an instance first from the builder to use it to create any part of the pipeline.

```go
builder := &pip.Builder[<TypeOfPipelineAndSteps>]{}
```

## Basic Step

Carries on a transformation on a single token in the pipeline and pushes it forward to the next steps.

```go
// The process type expected by the step
type StepBasicProcess[I any] func(I) I // Don't redefine

step := builder.NewStep(pip.StepBasicConfig[int64]{
    Label:    "minus10",
    Replicas: 3,
    // Process Is Required!
    Process:  func(token int64) int64 {
        return token - 10
    },
})
```

## Filter Step

Used to get rid of any undesired tokens from the pipelines.

```go
// StepFilterPassCriteria is function that determines if the data should be passed or not.
type StepFilterPassCriteria[I any] func(I) bool // Don't redefine

step := builder.NewStep(pip.StepFilterConfig[int64]{
    Replicas:     1,
    Label:        "filterEven",
    // PassCriteria Is Required!
    PassCriteria: func(token int64) bool {
        return token%2 == 0
    },
})
```

## Terminal Step

Terminal step is the final step in the pipeline and doesn't have an output. It is the process where you save the results of the pipeline to the database, send it over the network, ....

```go
// The process type expected by the terminal step.
type StepTerminalProcess[I any] func(I) // Don't redefine

// result step
resultStep := builder.NewStep(pip.StepTerminalConfig[int64]{
    Replicas: 1,
    Label:    "print",
    // Process Is Required!
    Process:  func(token int64) {
        fmt.Println("Result:", token)
    },
})
```

## Fragmenter Step (Example 5)

Fragmenter step allows users to break down a token into multiple tokens and fed to the following steps of the pipelines. This is useful when you are having a large chunk of data and want to breake down into a smaller problem and aggregate results later into a terminal step. You can reap its benifits of course when you increase the number of the replcias of subsequent steps.

```go
// The fragmenter process type expected by the fragmenter step.
type StepFragmenterProcess[I any] func(I) []I // Don't redefine

// fragmenter step
splitter := builder.NewStep(pip.StepFragmenterConfig[string]{
    Label:    "fragmenter",
    Replicas: 1,
    // Process Is Required!
    Process:  func(token string) []string {
        // Split the Value field by comma
        splitTokens := strings.Split(token.Value, ",")
        return splitTokens
    },
})
```

## Buffer Step (Examples 6 & 7)

Buffer step is a step that doesn't apply an operation on a single token, but buffers multiple items and act on them accordingly. The **collected tokens span** can be of 2 types:

1. Received Inputs Count Span
2. Periodic Time Span (The received token over a period of time)

The buffer step supports either or both of the 2 kinds of span. It also supports the pass through feature which allows users of the package to control whether the buffered token is sent to the following steps after being added to buffer or not.

You have also the option to flush the buffer or not after running your calculation on the buffer.

### Time Triggered Buffer Step (Example 6)

In the time triggered step, every input is added to the buffer, but the process **TimeTriggeredProcess** is called with the current buffer value every **TimeTriggeredProcessInterval** set in the step configuration.

In the following example, the step calculates the sum of the latest 5 elements (BufferSize) every 100 ms.

```go

// StepBufferProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferProcess[I any] func([]I) (I, BufferFlags) // Don't redefine

func periodicCalculateSum(buffer []int64) (int64, pip.BufferFlags) {
    var sum int64
    for _, v := range buffer {
        sum += v
    }
    return sum, pip.BufferFlags{
        SendProcessOuput:   true,
        FlushBuffer: false,
    }
}

bufferStep := builder.NewStep(pip.StepBufferConfig[int64]{
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
// StepBufferProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferProcess[I any] func([]I) (I, BufferFlags) // Don't redefine

func calculateSumOnBufferCountThreshold(buffer []int64) (int64, pip.BufferFlags) {

    // skipping the calcluation if not enough data was buffered.
    if len(buffer) < 10 {
        return 0, pip.BufferFlags{
            SendProcessOuput:   false,
            FlushBuffer:        false,
        }
    }

    var sum int64
    for _, v := range buffer {
        sum += v
    }
    return sum, pip.BufferFlags{
        SendProcessOuput:   true,
        FlushBuffer:        true, // This instructs the buffer step to flush the data it retains and start accumulating fresh data.
    }
}

bufferStep := builder.NewStep(pip.StepBufferConfig[int64]{
    Label:      "buffer",
    Replicas:   5,
    BufferSize: 10,
    PassThrough:            false,// This means the buffer is going to retain all elements.
    InputTriggeredProcess:  calculateSumOnBufferCountThreshold,
})
```

### Important Notes On Buffer Step

- The buffer size remains fixed, and when it is full the **oldest data is overwritten**. You will find the oldest data at index 0 and the most recent at the last element of the array.

- Although you can set replicas, but all operating on the concurrent safe buffer.

- The InputTriggered and the TimeTriggered processes are implemented in a concurrent safe context so the buffer will not change during their execution.

- When you set the **PassThrough** to true AND the input triggered process returns a valid result. The received input will be sent first to the following steps then the result from the process output.

- Again, you can set both time triggered and input triggered processes for the buffer step and they will be both be executed by their triggeres.

## Creating Custom Step

You can create an entirely different custom step by implementing the **IStep** interface methods.

Just take care that the following are provided by the pipeline and their values using setters, so don't set or run them yourself and implement only the setters. You can use them directly in the run method as needed.

- The input channel
- The output channel
- The decrement tokens handler
- The increment tokens handler
- Run method.

## Pipeline

Although the pipeline can operate on one type, but you can create a container structure to have a separate field for every step to set if you want to accumulate results of different types.

The pipeline creation requires 2 arguments

1. The configuration object of the pipeline.

2. Steps in the order of execution.

```go
config := pip.PipelineConfig{
    // The default step input channel size.
    // This is used for any step that has no input channel size explicitly set.
    DefaultStepInputChannelSize: 10,
    // Tells the pipeline to keep track of the tokens count.
    // This is important if you need to monitor the pipeline and want to use WaitTillDone function.
    // Setting this to false boosts the performance of steps since the tokens count is a critical section.
    TrackTokensCount:       true,
}
pipeline := builder.NewPipeline(config, step1, step2, step3, terminalStep)
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

WaitTillDone is used to block the execution till all the elements/tokens in the pipelines are processed. This requires some certain conditions to operate:

1. Requries pipeline confiugration **TrackTokensCount** to be set to true.

2. Requries terminal step to be used.

3. Requires every buffer to be flushed.

If these conditions are not met, **_WaitTillDone_** will stall the execution indefinitely

If you don't care about the current elements in the pipeline like in the case of data stream, you can call **Terminate()** directly without waiting.

```go
// Requries terminal step to be used.
// Requries pipeline confiugration TrackTokensCount to be set to true.
// Requires every buffer to be flushed.
pipeline.WaitTillDone()
```

It is an optional step to use and you can call Terminate() directly without waiting, but all the tokens in the pipeline will be discarded.

When using buffer step(s) don't use **pipeline.WaitTillDone()** unless you have a finite number of inputs and you flush the data in the buffer regularly. Otherwise wait till done will stall your application and may result a deadlock..

### Terminating Pipeline

When you want to terminate the pipeline use the following function. Note that it will terminate regardless the parent context is closed or not. And once it terminates, it can't be rerun again and you need to create another pipeline.

```go
pipeline.Terminate()
```
