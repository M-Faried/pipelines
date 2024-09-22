package pipelines

import (
	"context"
	"sync"
)

var once sync.Once

type IPipeline[I any] interface {
	// Init initializes the pipeline and has to be called once before running the pipeline.
	Init()
	// Run starts the pipeline.
	Run(ctx context.Context)
	// FeedOne feeds a single item to the pipeline.
	FeedOne(i I)
	// FeedMany feeds multiple items to the pipeline.
	FeedMany(i []I)
	// TokensCount returns the number of tokens being processed by the pipeline.
	TokensCount() uint64
	// WaitTillDone blocks until all tokens are consumed by the pipeline.
	WaitTillDone()
}

type pipeline[I any] struct {
	steps       []*Step[I]
	resultStep  *ResultStep[I]
	channelSize uint16

	tokensCount uint64
	tokensMutex sync.Mutex
	doneCond    *sync.Cond
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
	once.Do(func() {
		stepsCount := len(p.steps)
		channelsCount := stepsCount + 1

		// init channels
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

	// creating a wait group for all the step routines.
	wg := &sync.WaitGroup{}

	// running the result step.
	for range p.resultStep.replicas {
		wg.Add(1)
		go p.resultStep.run(ctx, wg)
	}

	// running steps in reverse order
	for i := len(p.steps) - 1; i >= 0; i-- {
		// spawning the replicas for each step
		for range p.steps[i].replicas {
			wg.Add(1)
			go p.steps[i].run(ctx, wg)
		}
	}

	// wait for the context to be done
	<-ctx.Done()

	// wait for step routines to be done
	wg.Wait()

	// closing all channels
	for _, step := range p.steps {
		close(step.input)
	}

	// closing the output of the last step which is the input to the result
	close(p.steps[len(p.steps)-1].output)
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

func (p *pipeline[I]) Append(other IPipeline[I]) {
	pOther := other.(*pipeline[I])
	// input for the other is the output for the current.
	pOther.steps[0].input = p.steps[len(p.steps)-1].output
	p.steps = append(p.steps, pOther.steps...)
}

func (p *pipeline[I]) TokensCount() uint64 {
	p.tokensMutex.Lock()
	defer p.tokensMutex.Unlock()
	return p.tokensCount
}

func (p *pipeline[I]) WaitTillDone() {
	p.doneCond.L.Lock()
	defer p.doneCond.L.Unlock()
	for p.tokensCount > 0 {
		p.doneCond.Wait()
	}
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
