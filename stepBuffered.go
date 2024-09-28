package pipelines

import (
	"context"
	"sync"
	"time"
)

type StepBufferedProcessOutput[I any] struct {
	HasResult bool
	Result    I
	Flush     bool
}

type StepBufferedTimeTiggeredProcess[I any] func([]I) StepBufferedProcessOutput[I]
type StepBufferedInputTiggeredProcess[I any] func([]I) StepBufferedProcessOutput[I]

type StepBuffered[I any] struct {
	Label       string
	Replicas    uint16
	BufferSize  int
	PassThrough bool

	InputTriggeredProcess        StepBufferedInputTiggeredProcess[I]
	TimeTriggeredProcess         StepBufferedTimeTiggeredProcess[I]
	TimeTriggeredProcessInterval time.Duration
}

type stepBuffered[I any] struct {
	stepBase[I]
	buffer      []I
	bufferSize  int
	bufferMutex sync.Mutex
	passThrough bool

	inputTriggeredProcess        StepBufferedInputTiggeredProcess[I]
	timeTriggeredProcess         StepBufferedTimeTiggeredProcess[I]
	timeTriggeredProcessInterval time.Duration
}

func (s *stepBuffered[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	// s.initBuffer()
	if s.timeTriggeredProcessInterval == 0 {
		// 10 hours is to cover the case where the time is not set. The buffer in this case will be input triggered.
		// this is needed to keep the thread alive as well inc case there is a long delay in the input.
		s.timeTriggeredProcessInterval = 1000 * time.Hour
	}
	timer := time.NewTimer(s.timeTriggeredProcessInterval)
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case i, ok := <-s.input:
			if !ok {
				wg.Done()
				return
			}
			// s.addToBuffer(i)
			s.handleInputTriggeredProcess(i)
			timer.Reset(s.timeTriggeredProcessInterval)
		case <-timer.C:
			s.handleTimeTriggeredProcess()
			timer.Reset(s.timeTriggeredProcessInterval)
		}
	}
}

func (s *stepBuffered[I]) handleInputTriggeredProcess(i I) {

	// All the following has to be done in during the same mutex lock.
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	// Adding the input to buffer.
	if len(s.buffer) == s.bufferSize {
		s.buffer = s.buffer[1:]
	}
	s.buffer = append(s.buffer, i)

	// Checking if the passThrough is set and passing the input if it is.
	if s.passThrough {
		// since it was already stored the token in the buffer and will be passed in the next steps,
		// we need to increment the tokens count by one.
		s.incrementTokensCount()
		s.output <- i
	}

	// Checking if the input triggered process is set.
	if s.inputTriggeredProcess == nil {
		return
	}

	// Processing the buffer after adding the element.
	processOutput := s.inputTriggeredProcess(s.buffer)

	if processOutput.HasResult {
		// Since this is a new result, we need to increment the tokens count.
		s.incrementTokensCount()
		s.output <- processOutput.Result
	}

	// Check if the buffer should be flushed or not.
	if processOutput.Flush {
		length := len(s.buffer)
		for range length {
			s.decrementTokensCount()
		}
		s.buffer = s.buffer[:0]
	}
}

func (s *stepBuffered[I]) handleTimeTriggeredProcess() {

	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	processOutput := s.timeTriggeredProcess(s.buffer)

	if processOutput.HasResult {
		s.incrementTokensCount()
		s.output <- processOutput.Result
	}

	// Check if the buffer should be flushed or not.
	if processOutput.Flush {
		length := len(s.buffer)
		for range length {
			s.decrementTokensCount()
		}
		s.buffer = s.buffer[:0]
	}
}
