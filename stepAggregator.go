package pipelines

import (
	"context"
	"sync"
	"time"
)

// StepAggregatorThresholdCriteria is a function that determines if the threshold for the aggregation result
// to be calculated is reached (based on data) or not.
type StepAggregatorThresholdCriteria[I any] func([]I) bool

// StepAggregatorProcess is a function that processes the buffered data inside the aggregator step.
type StepAggregatorProcess[I any] func([]I) (I, error)

// StepAggregatorConfig is a struct that defines the configuration for an aggregator step. The aggregator step accumulates data matching a certain
// criteria and then when this data reaches a sepcified threshold, or a certain time has passed, the data is processed.
type StepAggregatorConfig[I any] struct {
	Label               string
	Replicas            uint16
	ErrorHandler        ErrorHandler
	Process             StepAggregatorProcess[I]
	ThresholdCriteria   StepAggregatorThresholdCriteria[I]
	AggregationInterval time.Duration
}

// stepAggregator is a struct that defines an aggregator step.
type stepAggregator[I any] struct {
	stepBase[I]

	process             StepAggregatorProcess[I]
	thresholdCriteria   StepAggregatorThresholdCriteria[I]
	aggregationInterval time.Duration

	buffer      []I
	bufferMutex sync.Mutex
}

// addToBuffer adds an element to the buffer.
func (s *stepAggregator[I]) addToBuffer(i I) {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	s.buffer = append(s.buffer, i)
}

// isThresholdReached checks if the data hit the threshold specified by the user or not.
func (s *stepAggregator[I]) isThresholdReached() bool {
	if s.thresholdCriteria == nil {
		return false
	}
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	return s.thresholdCriteria(s.buffer)
}

// processBufferIfNotEmpty applies the processes to the buffer of the aggregator step.
func (s *stepAggregator[I]) processBufferIfNotEmpty() {

	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	// Returning if there are no data in the buffer
	if len(s.buffer) == 0 {
		return
	}

	// decrementing the tokens count for each element in the buffer except the last one
	// to count for the result token from the aggregator step.
	for range len(s.buffer) - 1 {
		s.decrementTokensCount()
	}
	// processing the buffer
	item, err := s.process(s.buffer)
	if err != nil {
		s.reportError(err)
		// the whole result of the aggregation is filtered out so remove the aggregation result token.
		s.decrementTokensCount()
	} else {
		s.output <- item
	}
	s.buffer = make([]I, 0)
}

func (s *stepAggregator[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	if s.aggregationInterval == 0 {
		// 10 hours is to cover the case where the time is not set. The threshould should be met far before this time and the timer is reset.
		// this is needed to keep the thread alive as well inc case there is a long delay in the input.
		s.aggregationInterval = 100 * time.Hour
	}
	timer := time.NewTimer(s.aggregationInterval)
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
			// if the element is matching the criteria we add it to the buffer.
			s.addToBuffer(i)
			// if the buffer reached the required threshould we process it
			if s.isThresholdReached() {
				s.processBufferIfNotEmpty()
				timer.Reset(s.aggregationInterval)
			}
		case <-timer.C:
			s.processBufferIfNotEmpty()
			timer.Reset(s.aggregationInterval)
		}
	}
}
