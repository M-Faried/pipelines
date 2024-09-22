package pipelines

import (
	"context"
	"sync"
)

var initOnce sync.Once
var runOnce sync.Once

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

type pipeline[I any] struct {

	// steps is the list of steps in the pipeline.
	steps []*Step[I]

	// resultStep is the final step in the pipeline.
	resultStep *ResultStep[I]

	// channelSize is the buffer size for all channels used to connect steps.
	channelSize uint16

	// stepsCtxCancel is used to cancel the context of the steps.
	stepsCtxCancel context.CancelFunc

	// stepsWaitGroup is used to wait for all the step routines to receive ctx cancel signal.
	stepsWaitGroup *sync.WaitGroup

	// tokensCount is the number of tokens being processed by the pipeline.
	tokensCount uint64

	// tokensMutex is protect tokensCount from race conditions.
	tokensMutex sync.Mutex

	// doneCond is used to wait till all tokens are processed.
	doneCond *sync.Cond
}

// NewPipeline creates a new pipeline with the given channel size, result step and steps.
// The channel size is the buffer size for all channels used to connect steps.
func NewPipeline[I any](channelSize uint16, resultStep *ResultStep[I], steps ...*Step[I]) IPipeline[I] {
	pipe := &pipeline[I]{}
	pipe.steps = steps
	pipe.resultStep = resultStep
	pipe.channelSize = channelSize
	pipe.doneCond = sync.NewCond(&pipe.tokensMutex)
	return pipe
}

func (p *pipeline[I]) Init() {
	initOnce.Do(func() {

		// getting the number of steps and channels
		stepsCount := len(p.steps)
		channelsCount := stepsCount + 1

		// creating the required channels
		allChannels := make([]chan I, channelsCount)
		for i := 0; i < channelsCount; i++ {
			allChannels[i] = make(chan I, p.channelSize)
		}

		// setting channels for each step
		for i := 0; i < stepsCount && (i+1) < channelsCount; i++ {
			p.steps[i].input = allChannels[i]
			p.steps[i].output = allChannels[i+1]
			p.steps[i].decrementTokensCount = p.decrementTokensCount
		}

		// setting the input for the result step.
		p.resultStep.input = p.steps[stepsCount-1].output
		p.resultStep.decrementTokensCount = p.decrementTokensCount
	})
}

func (p *pipeline[I]) Run(ctx context.Context) {
	runOnce.Do(func() {
		// creating a wait group for all the step routines.
		p.stepsWaitGroup = &sync.WaitGroup{}

		stepsCtx, cancel := context.WithCancel(ctx)
		p.stepsCtxCancel = cancel

		// running the result step.
		for range p.resultStep.replicas {
			p.stepsWaitGroup.Add(1)
			go p.resultStep.run(stepsCtx, p.stepsWaitGroup)
		}

		// running steps in reverse order
		for i := len(p.steps) - 1; i >= 0; i-- {
			// spawning the replicas for each step
			for range p.steps[i].replicas {
				p.stepsWaitGroup.Add(1)
				go p.steps[i].run(stepsCtx, p.stepsWaitGroup)
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

	// canceling the context in case the parent context is not cancelled.
	p.stepsCtxCancel()

	// wait for step routines to be done
	p.stepsWaitGroup.Wait()

	// closing all channels
	for _, step := range p.steps {
		close(step.input)
	}

	// closing the output of the last step which is the input to the result
	close(p.steps[len(p.steps)-1].output)

	// clearing the wait group
	p.stepsWaitGroup = nil
}

func (p *pipeline[I]) FeedOne(item I) {
	p.incrementTokensCount()
	p.steps[0].input <- item
}

func (p *pipeline[I]) FeedMany(items []I) {
	for _, item := range items {
		p.incrementTokensCount()
		p.steps[0].input <- item
	}
}

func (p *pipeline[I]) TokensCount() uint64 {
	p.tokensMutex.Lock()
	defer p.tokensMutex.Unlock()
	return p.tokensCount
}

func (p *pipeline[I]) incrementTokensCount() {
	p.tokensMutex.Lock()
	defer p.tokensMutex.Unlock()
	p.tokensCount++
}

func (p *pipeline[I]) decrementTokensCount() {
	p.tokensMutex.Lock()
	defer p.tokensMutex.Unlock()
	p.tokensCount--
	p.doneCond.Signal()
}
