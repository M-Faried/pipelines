package pipelines

import (
	"context"
	"sync"
	"time"
)

// StepBufferedProcessOutput is the output of the StepBufferedProcess.
type StepBufferedProcessOutput[I any] struct {

	// HasResult is a flag to indicate if the process has a result or not.
	HasResult bool

	// Result is the result of the process.
	Result I

	// FlushBuffer signals the step to flush all the values stored in the buffer.
	FlushBuffer bool
}

// StepBufferedProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferedProcess[I any] func([]I) StepBufferedProcessOutput[I]

type StepBuffered[I any] struct {

	// Label is the name of the step.
	Label string

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// BufferSize is the max size of the buffer. If the buffer is full, the oldest element will be removed.
	BufferSize int

	// PassThrough is a flag to pass the input to the following step or not in addition to adding it to the buffer.
	// If PassThrough is set to false, the buffer will retain all the elements in the buffer.
	PassThrough bool

	// The process which is called when the input is received and added to the buffer.
	InputTriggeredProcess StepBufferedProcess[I]

	// The process which is called periodically based on the interval.
	TimeTriggeredProcess StepBufferedProcess[I]

	// TimeTriggeredProcessInterval is the interval at which the TimeTriggeredProcess is called.
	TimeTriggeredProcessInterval time.Duration
}

type stepBuffered[I any] struct {
	stepBase[I]
	buffer      []I
	bufferSize  int
	bufferMutex sync.Mutex
	passThrough bool

	inputTriggeredProcess        StepBufferedProcess[I]
	timeTriggeredProcess         StepBufferedProcess[I]
	timeTriggeredProcessInterval time.Duration
}

func (s *stepBuffered[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	if s.timeTriggeredProcessInterval == 0 {
		// 1000 hours is to cover the case where the time is not set. The buffer in this case will be input triggered.
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
			s.handleInputTriggeredProcess(i)
			timer.Reset(s.timeTriggeredProcessInterval)
		case <-timer.C:
			s.handleTimeTriggeredProcess()
			timer.Reset(s.timeTriggeredProcessInterval)
		}
	}
}

func (s *stepBuffered[I]) addToBuffer(i I) bool {
	overwriteOccurred := false
	if len(s.buffer) == s.bufferSize {
		s.buffer = s.buffer[1:]
		overwriteOccurred = true
	}
	s.buffer = append(s.buffer, i)
	return overwriteOccurred
}

func (s *stepBuffered[I]) handleInputTriggeredProcess(i I) {

	// All the following has to be done in during the same mutex lock.
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	// Adding the input to buffer.
	overwriteOccurred := s.addToBuffer(i)

	// Checking if the passThrough is set and passing the input if it is.
	if s.passThrough {
		// since it was already stored the token in the buffer and will be passed in the next steps,
		// we need to increment the tokens count by one if it is stored in the buffer.
		// In other words, the max increments are going to be the buffer size.
		if !overwriteOccurred {
			s.incrementTokensCount()
		}
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
	if processOutput.FlushBuffer {
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
	if processOutput.FlushBuffer {
		length := len(s.buffer)
		for range length {
			s.decrementTokensCount()
		}
		s.buffer = s.buffer[:0]
	}
}
