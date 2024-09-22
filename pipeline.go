package pipelines

import (
	"context"
	"sync"
)

var once sync.Once

type IPipeline[I any] interface {
	Init()
	Run(ctx context.Context)
	FeedOne(i I)
	FeedMany(i []I)
	ResultChannel() <-chan I
	ReadError() error
	Append(p IPipeline[I])
}

type pipeline[I any] struct {
	steps       []*Step[I]
	resultStep  *ResultStep[I]
	errorsQueue *Queue[error]
	channelSize uint16
}

func (p *pipeline[I]) Init() {
	once.Do(func() {
		stepsCount := len(p.steps)
		channelsCount := stepsCount + 1

		// init error queue
		p.errorsQueue = NewQueue[error]()

		// init channels
		allChannels := make([]chan I, channelsCount)
		for i := 0; i < channelsCount; i++ {
			allChannels[i] = make(chan I, p.channelSize)
		}

		// init steps
		for i := 0; i < stepsCount && (i+1) < channelsCount; i++ {
			p.steps[i].input = allChannels[i]
			p.steps[i].output = allChannels[i+1]
			p.steps[i].errorsQueue = p.errorsQueue
		}

		// init result handler if exists
		if p.resultStep == nil {
			return
		}
		p.resultStep.input = p.steps[stepsCount-1].output
	})
}

func (p *pipeline[I]) Run(ctx context.Context) {

	// creating a wait group for all the step routines.
	wg := &sync.WaitGroup{}

	// running the result step if it exists
	if p.resultStep != nil {
		for range p.resultStep.replicas {
			wg.Add(1)
			go p.resultStep.run(ctx, wg)
		}
	}

	// running steps in reverse order
	for i := len(p.steps) - 1; i >= 0; i-- {
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
	// closing the output of the last step which is the input to the result if it exists
	close(p.steps[len(p.steps)-1].output)
}

func (p *pipeline[I]) FeedOne(item I) {
	p.steps[0].input <- item
}

func (p *pipeline[I]) FeedMany(items []I) {
	for _, item := range items {
		p.steps[0].input <- item
	}
}

func (p *pipeline[I]) ReadError() error {
	return p.errorsQueue.Dequeue()
}

func (p *pipeline[I]) ResultChannel() <-chan I {
	lastStepIndex := len(p.steps) - 1
	resultChannel := p.steps[lastStepIndex].output
	return resultChannel
}

func (p *pipeline[I]) Append(other IPipeline[I]) {
	pOther := other.(*pipeline[I])
	// input for the other is the output for the current.
	pOther.steps[0].input = p.steps[len(p.steps)-1].output
	p.steps = append(p.steps, pOther.steps...)
}
