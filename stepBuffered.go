package pipelines

import (
	"context"
	"sync"
	"time"
)

type TimeTriggeredProcessOutput[I any] struct {
	HasResult bool
	Result    I
	Flush     bool
}

type InputTriggeredProcessOutput[I any] struct {
	HasResult     bool
	Result        I
	Flush         bool
	PassSameToken bool
}

type StepBufferedTimeTiggeredProcess[I any] func([]I) TimeTriggeredProcessOutput[I]
type StepBufferedInputTiggeredProcess[I any] func([]I) InputTriggeredProcessOutput[I]

type StepBuffered[I any] struct {
	Label      string
	Replicas   uint16
	BufferSize int

	InputTriggeredProcess        StepBufferedInputTiggeredProcess[I]
	TimeTriggeredProcess         StepBufferedTimeTiggeredProcess[I]
	TimeTriggeredProcessInterval time.Duration
}

type stepBuffered[I any] struct {
	stepBase[I]
	buffer      []I
	bufferSize  int
	bufferMutex sync.Mutex
	// bufferTokensCount int

	inputTriggeredProcess        StepBufferedInputTiggeredProcess[I]
	timeTriggeredProcess         StepBufferedTimeTiggeredProcess[I]
	timeTriggeredProcessInterval time.Duration
}

// func (s *stepBuffered[I]) addToBuffer(i I) {
// 	s.bufferMutex.Lock()
// 	defer s.bufferMutex.Unlock()
// 	// this means the newes element will always be at the end of the
// 	if len(s.buffer) == s.bufferSize {
// 		s.buffer = s.buffer[1:]
// 	}
// 	s.buffer = append(s.buffer, i)

// }

// func (s *stepBuffered[I]) initBuffer() {
// 	s.bufferMutex.Lock()
// 	defer s.bufferMutex.Unlock()
// 	s.buffer = make([]I, 0, s.bufferSize)
// }

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
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	// Adding to buffer
	if len(s.buffer) == s.bufferSize {
		s.buffer = s.buffer[1:]
	}
	s.buffer = append(s.buffer, i)

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

	if processOutput.PassSameToken {
		// since it was already stored in the buffer and will be passed in the next step,
		// we need to increment the tokens count.
		s.incrementTokensCount()
		s.output <- i
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
