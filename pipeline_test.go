package pipelines

import (
	"context"
	"testing"
	"time"
)

func TestPipeline_Init(t *testing.T) {

	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}

	p.Init()

	if p.steps[0].GetInputChannel() == nil {
		t.Errorf("expected input channel to be initialized")
	}
	if p.steps[0].GetOutputChannel() == nil {
		t.Errorf("expected output channel to be initialized")
	}
	if p.steps[1].GetInputChannel() == nil {
		t.Errorf("expected output channel to be initialized")
	}
	if p.doneCond == nil {
		t.Errorf("expected doneCond to be initialized")
	}
	if p.tokensCount != 0 {
		t.Errorf("expected tokens count to be 0, got %d", p.tokensCount)
	}

	p.WaitTillDone()
	p.Terminate()
}

func TestPipeline_Run(t *testing.T) {
	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}
	p.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Run(ctx)

	if p.stepsCtxCancel == nil {
		t.Errorf("expected stepsCtxCancel to be initialized")
	}

	if p.stepsWaitGroup == nil {
		t.Errorf("expected stepsWaitGroup to be initialized")
	}

	p.WaitTillDone()
	p.Terminate()
}

func TestPipeline_FeedOne(t *testing.T) {
	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}
	p.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p.Run(ctx)

	p.FeedOne(1)
	if p.TokensCount() != 1 {
		t.Errorf("expected tokens count to be 1, got %d", p.TokensCount())
	}

	time.Sleep(100 * time.Millisecond) // wait for the pipeline to process the items

	if p.TokensCount() != 0 {
		t.Errorf("expected tokens count to be 0, got %d", p.TokensCount())
	}

	p.WaitTillDone()
	p.Terminate()
}

func TestPipeline_FeedMany(t *testing.T) {
	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}
	p.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p.Run(ctx)

	p.FeedMany([]int{1, 2, 3})

	if p.TokensCount() != 3 {
		t.Errorf("expected tokens count to be 3, got %d", p.TokensCount())
	}

	time.Sleep(100 * time.Millisecond) // wait for the pipeline to process the items

	if p.TokensCount() != 0 {
		t.Errorf("expected tokens count to be 0, got %d", p.TokensCount())
	}

	p.WaitTillDone()
	p.Terminate()
}

func TestPipeline_Terminate(t *testing.T) {
	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}
	p.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p.Init()
	p.Run(ctx)
	p.Terminate()

	if p.stepsWaitGroup != nil {
		t.Errorf("expected stepsWaitGroup to be nil after termination")
	}

	// Check if the input channel of the first step is closed
	select {
	case _, ok := <-p.steps[0].GetInputChannel():
		if ok {
			t.Errorf("expected input channel of the first step to be closed")
		}
	case _, ok := <-p.steps[1].GetInputChannel():
		if ok {
			t.Errorf("expected input channel of the second step to be closed")
		}
	}
}

func TestPipeline_WaitTillDone(t *testing.T) {
	steps := []iStepInternal[int]{
		&mockStep[int]{replicas: 1},
		&mockStep[int]{replicas: 1, finalStep: true},
	}

	p := &pipeline[int]{
		steps:       steps,
		channelSize: 10,
	}
	p.Init()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	p.Run(ctx)

	p.FeedMany([]int{1, 2, 3})
	p.WaitTillDone()

	if p.TokensCount() != 0 {
		t.Errorf("expected tokens count to be 0, got %d", p.TokensCount())
	}
	p.Terminate()
}