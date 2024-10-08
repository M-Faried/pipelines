package pipelines

import (
	"context"
	"sync"
	"time"
)

// BufferFlags is flags instructing the buffer step to behave in certain way.
type BufferFlags struct {

	// SendProcessOuput is a flag to indicate if the process has a result or not.
	SendProcessOuput bool

	// FlushBuffer signals the step to flush all the values stored in the buffer.
	FlushBuffer bool
}

// StepBufferProcess is the function signature for the process which is called periodically or when the input is received.
type StepBufferProcess[I any] func([]I) (I, BufferFlags)

// StepBufferConfig is the confiuration for creating a buffer step.
type StepBufferConfig[I any] struct {

	// Label is the name of the step.
	Label string

	// Replicas is the number of replicas (go routines) created to run the step.
	Replicas uint16

	// InputChannelSize is the buffer size for the input channel to the step
	InputChannelSize uint16

	// BufferSize is the max size of the buffer. If the buffer is full, the oldest element will be removed.
	BufferSize int

	// PassThrough is a flag to pass the input to the following step or not in addition to adding it to the buffer.
	// If PassThrough is set to false, the buffer will retain all the elements in the buffer.
	PassThrough bool

	// The process which is called when the input is received and added to the buffer.
	InputTriggeredProcess StepBufferProcess[I]

	// The process which is called periodically based on the interval.
	TimeTriggeredProcess StepBufferProcess[I]

	// TimeTriggeredProcessInterval is the interval at which the TimeTriggeredProcess is called.
	TimeTriggeredProcessInterval time.Duration
}

type stepBuffer[I any] struct {
	stepBase[I]
	buffer      []I
	bufferSize  int
	bufferMutex sync.Mutex
	passThrough bool

	inputTriggeredProcess        StepBufferProcess[I]
	timeTriggeredProcess         StepBufferProcess[I]
	timeTriggeredProcessInterval time.Duration
}

func newStepBuffer[I any](config StepBufferConfig[I]) IStep[I] {
	if config.InputTriggeredProcess == nil && config.TimeTriggeredProcess == nil {
		panic("either time triggered or input process is required")
	}
	if config.TimeTriggeredProcess != nil && config.TimeTriggeredProcessInterval == 0 {
		panic("time triggered process interval is required to be used with time triggered process")
	}
	if config.BufferSize <= 0 {
		panic("buffer size must be greater than or equal to 0")
	}

	return &stepBuffer[I]{
		stepBase:                     newBaseStep[I](config.Label, config.Replicas, config.InputChannelSize),
		bufferSize:                   config.BufferSize,
		passThrough:                  config.PassThrough,
		buffer:                       make([]I, 0, config.BufferSize),
		inputTriggeredProcess:        config.InputTriggeredProcess,
		timeTriggeredProcess:         config.TimeTriggeredProcess,
		timeTriggeredProcessInterval: config.TimeTriggeredProcessInterval,
	}
}

func (s *stepBuffer[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	if s.timeTriggeredProcessInterval == 0 {
		// 1000 hours is to cover the case where the time is not set. The buffer in this case will be input triggered.
		// this is needed to keep the thread alive as well inc case there is a long delay in the input.
		s.timeTriggeredProcessInterval = 1000 * time.Hour
	}
	ticker := time.NewTicker(s.timeTriggeredProcessInterval)
	defer ticker.Stop()
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case i, ok := <-s.input:
			if !ok {
				return
			}
			s.handleInputTriggeredProcess(i)
		case <-ticker.C:
			s.handleTimeTriggeredProcess()
		}
	}
}

func (s *stepBuffer[I]) addToBuffer(i I) bool {
	overwriteOccurred := false
	if len(s.buffer) == s.bufferSize {
		s.buffer = s.buffer[1:]
		overwriteOccurred = true
	}
	s.buffer = append(s.buffer, i)
	return overwriteOccurred
}

func (s *stepBuffer[I]) handleInputTriggeredProcess(i I) {

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
	processOutput, flags := s.inputTriggeredProcess(s.buffer)

	if flags.SendProcessOuput {
		// Since this is a new result, we need to increment the tokens count.
		s.incrementTokensCount()
		s.output <- processOutput
	}

	// Check if the buffer should be flushed or not.
	if flags.FlushBuffer {
		length := len(s.buffer)
		for range length {
			s.decrementTokensCount()
		}
		s.buffer = s.buffer[:0]
	}
}

func (s *stepBuffer[I]) handleTimeTriggeredProcess() {

	if s.timeTriggeredProcess == nil {
		return
	}

	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	processOutput, flags := s.timeTriggeredProcess(s.buffer)

	// Check if the process has a result or not.
	if flags.SendProcessOuput {
		s.incrementTokensCount()
		s.output <- processOutput
	}

	// Check if the buffer should be flushed or not.
	if flags.FlushBuffer {
		length := len(s.buffer)
		for range length {
			s.decrementTokensCount()
		}
		s.buffer = s.buffer[:0]
	}
}
