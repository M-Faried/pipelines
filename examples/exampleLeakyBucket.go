package examples

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	pip "github.com/m-faried/pipelines"
)

const TEXT = "piplines is an awesome package !!!"

type leakyBucket struct {
	addedTokensCount int
	mutex            sync.Mutex
}

func (l *leakyBucket) inputTriggered(buffer []string) (string, pip.BufferFlags) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.addedTokensCount < len(buffer) {
		l.addedTokensCount++
	}
	return "", pip.BufferFlags{}
}

func (l *leakyBucket) timeTriggeredProcess(buffer []string) (string, pip.BufferFlags) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	itemIndex := len(buffer) - l.addedTokensCount

	if l.addedTokensCount > 1 {
		l.addedTokensCount--
	} else if l.addedTokensCount == 1 {
		l.addedTokensCount = len(buffer)
	}

	return buffer[itemIndex], pip.BufferFlags{SendProcessOuput: true}
}

func createLeakyBucketPipeline() pip.IPipeline[string] {
	builder := &pip.Builder[string]{}

	wordSplitter := builder.NewStep(pip.StepFragmenterConfig[string]{
		Label: "word splitter",
		Process: func(in string) []string {
			return strings.Split(in, " ")
		},
	})

	wordTrimmer := builder.NewStep(pip.StepBasicConfig[string]{
		Label: "word trimmer",
		Process: func(in string) string {
			return strings.Trim(in, " ,;:")
		},
	})

	wordFilter := builder.NewStep(pip.StepFilterConfig[string]{
		Label: "filter",
		PassCriteria: func(in string) bool {
			return in != ""
		},
	})

	leakyBucket := &leakyBucket{}
	bucketStep := builder.NewStep(pip.StepBufferConfig[string]{
		Label:                        "Leaky Bucket Step",
		BufferSize:                   10,
		InputTriggeredProcess:        leakyBucket.inputTriggered,
		TimeTriggeredProcess:         leakyBucket.timeTriggeredProcess,
		TimeTriggeredProcessInterval: 1 * time.Second,
	})

	printer := builder.NewStep(pip.StepTerminalConfig[string]{
		Label: "terminal",
		Process: func(in string) {
			fmt.Println(in)
		},
	})

	pConfig := pip.PipelineConfig{
		DefaultChannelSize: 10,
		TrackTokensCount:   false,
	}

	pipeline := builder.NewPipeline(pConfig, wordSplitter, wordTrimmer, wordFilter, bucketStep, printer)
	pipeline.Init()
	return pipeline
}

func ExampleLeakyBucket() {

	// Set up a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancelCtx := context.WithCancel(context.Background())

	pipeline := createLeakyBucketPipeline()
	pipeline.Run(ctx)
	pipeline.FeedOne(TEXT)

	// Goroutine to cancel the context when an interrupt signal is received
	go func() {
		<-sigChan
		fmt.Println("Interrupt signal received, cancelling context...")

		// cancel the context
		cancelCtx()

		// terminating the pipeline and clearning resources
		pipeline.Terminate()

		fmt.Println("Pipeline Terminated!!!")
	}()

	<-ctx.Done()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Example Leaky Bucket Done!!!")
}
