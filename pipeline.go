package pipelines

import (
	"context"
	"sync"
)

// IPipeline is an interface that represents a pipeline.
type IPipeline[I any] interface {

	// Init initializes the pipeline and has to be called once before running the pipeline.
	Init()

	// Run starts the pipeline.
	Run(ctx context.Context)

	// WaitTillDone blocks until all tokens are consumed by the pipeline.
	WaitTillDone()

	// Terminate blocks and closes all the channels
	Terminate()

	// FeedOne feeds a single item to the pipeline.
	FeedOne(i I)

	// FeedMany feeds multiple items to the pipeline.
	FeedMany(i []I)

	// TokensCount returns the number of tokens being processed by the pipeline.
	TokensCount() uint64
}

// pipeline is a struct that represents a pipeline.
type pipeline[I any] struct {

	// steps is the list of steps in the pipeline.
	steps []iStepInternal[I]

	// channelSize is the buffer size for all channels used to connect steps.
	channelSize uint16

	// stepsCtxCancel is used to cancel the context of the steps.
	stepsCtxCancel context.CancelFunc

	// stepsWaitGroup is used to wait for all the step routines to receive ctx cancel signal.
	stepsWaitGroup *sync.WaitGroup

	// tokensCount is the number of tokens being processed by the pipeline.
	tokensCount uint64

	// tokensCountMutex is protect tokensCount from race conditions.
	tokensCountMutex sync.Mutex

	// doneCond is used to wait till all tokens are processed.
	doneCond *sync.Cond

	// initOnce is used to initialize the pipeline only once.
	initOnce sync.Once

	// runOnce is used to run the pipeline only once.
	runOnce sync.Once

	// channelsClosed is used to signal that all channels are closed.
	channelsClosed bool
}

func (p *pipeline[I]) Init() {
	p.initOnce.Do(func() {
		// creating a condition variable for the done condition
		p.doneCond = sync.NewCond(&p.tokensCountMutex)

		// getting the number of steps and channels
		stepsCount := len(p.steps)

		// creating the required channels
		// an input for each step
		allChannels := make([]chan I, stepsCount)
		for i := 0; i < stepsCount; i++ {
			allChannels[i] = make(chan I, p.channelSize)
		}

		// setting channels for each step
		for i := 0; i < stepsCount-1; i++ {
			p.steps[i].SetInputChannel(allChannels[i])
			p.steps[i].SetOutputChannel(allChannels[i+1])
			// setting decrement in case of filtering occurs at the step
			p.steps[i].SetDecrementTokensCountHandler(p.decrementTokensCount)
			// setting increment in case of fragmentation occurs at the step
			p.steps[i].SetIncrementTokensCountHandler(p.incrementTokensCount)
		}

		// setting the input for the result step.
		stepBeforeResultIndex := stepsCount - 2
		resultStepIndex := stepsCount - 1
		// setting the input for the result step from the step before it.
		p.steps[resultStepIndex].SetInputChannel(p.steps[stepBeforeResultIndex].GetOutputChannel())
		// setting the decrement for the result step.
		p.steps[resultStepIndex].SetDecrementTokensCountHandler(p.decrementTokensCount)
	})
}

func (p *pipeline[I]) Run(ctx context.Context) {
	p.runOnce.Do(func() {
		// creating a wait group for all the step routines.
		p.stepsWaitGroup = &sync.WaitGroup{}

		// creating a child context for the steps from the parent context.
		stepsCtx, cancel := context.WithCancel(ctx)
		p.stepsCtxCancel = cancel

		// running steps in reverse order
		for i := len(p.steps) - 1; i >= 0; i-- {
			// spawning the replicas for each step
			for range p.steps[i].GetReplicas() {
				p.stepsWaitGroup.Add(1)
				go p.steps[i].Run(stepsCtx, p.stepsWaitGroup)
			}
		}
	})
}

func (p *pipeline[I]) WaitTillDone() {
	p.doneCond.L.Lock()
	defer p.doneCond.L.Unlock()
	for p.tokensCount > 0 {
		p.doneCond.Wait()
	}
}

func (p *pipeline[I]) Terminate() {

	// checking the steps are running
	if p.stepsWaitGroup == nil {
		return
	}

	// setting closed channel signal immediately so that the input stops.
	p.channelsClosed = true

	// canceling the context in case the parent context is not cancelled.
	p.stepsCtxCancel()

	// wait for step routines to be done
	p.stepsWaitGroup.Wait()

	// closing all channels
	for _, step := range p.steps {
		close(step.GetInputChannel())
	}

	// clearing the wait group
	p.stepsWaitGroup = nil
}

func (p *pipeline[I]) FeedOne(item I) {
	if p.channelsClosed {
		return
	}
	p.incrementTokensCount()
	p.steps[0].GetInputChannel() <- item
}

func (p *pipeline[I]) FeedMany(items []I) {
	for _, item := range items {
		if p.channelsClosed {
			return
		}
		p.incrementTokensCount()
		p.steps[0].GetInputChannel() <- item
	}
}

func (p *pipeline[I]) TokensCount() uint64 {
	p.tokensCountMutex.Lock()
	defer p.tokensCountMutex.Unlock()
	return p.tokensCount
}

func (p *pipeline[I]) incrementTokensCount() {
	p.tokensCountMutex.Lock()
	defer p.tokensCountMutex.Unlock()
	p.tokensCount++
}

func (p *pipeline[I]) decrementTokensCount() {
	p.tokensCountMutex.Lock()
	defer p.tokensCountMutex.Unlock()
	p.tokensCount--
	p.doneCond.Signal()
}
