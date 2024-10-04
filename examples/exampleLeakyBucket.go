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

const (
	TEXT          = "piplines is an awesome package !!!"
	TEXT_2        = "New Text, 2 Available"
	TEXT_3        = "New Text 3 Arrived:"
	TEXT_4        = "This is the last Text"
	BUFFER_LENGTH = 10
)

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

	if l.addedTokensCount > 0 {
		l.addedTokensCount--
	}

	if itemIndex < len(buffer) {
		return buffer[itemIndex], pip.BufferFlags{SendProcessOuput: true}
	} else {
		return "----", pip.BufferFlags{SendProcessOuput: true}
	}

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
		BufferSize:                   BUFFER_LENGTH,
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
		DefaultStepChannelSize: 10,
		TrackTokensCount:       false,
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

	go func() {
		pipeline.FeedOne(TEXT)
		time.Sleep(2 * time.Second)
		pipeline.FeedOne(TEXT_2)
		time.Sleep(15 * time.Second)
		pipeline.FeedOne(TEXT_3)
		time.Sleep(15 * time.Second)
		pipeline.FeedOne(TEXT_4)
	}()

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
