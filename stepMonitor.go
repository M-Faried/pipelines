package pipelines

import (
	"context"
	"sync"
	"time"
)

// StepMonitorNotifyCriteria is a function that determines whether or not the listener should be notified.
type StepMonitorNotifyCriteria[I any] func([]I) bool

// StepMonitorNotify is a function that notifies the listener.
type StepMonitorNotify[I any] func([]I)

type StepMonitorConfig[I any] struct {
	Label          string
	Replicas       uint16
	NotifyCriteria StepMonitorNotifyCriteria[I]
	Notify         StepMonitorNotify[I]
	CheckInterval  time.Duration
}

// stepAggregator is a struct that defines an aggregator step.
type stepMonitor[I any] struct {
	stepBase[I]

	notifyCriteria StepMonitorNotifyCriteria[I]
	notify         StepMonitorNotify[I]
	checkInterval  time.Duration

	buffer      []I
	bufferMutex sync.Mutex
}

// addToBuffer adds an element to the buffer.
func (s *stepMonitor[I]) addToBuffer(i I) {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	s.buffer = append(s.buffer, i)
}

// isThresholdReached checks if the data hit the threshold specified by the user or not.
func (s *stepMonitor[I]) shouldNotify() bool {
	if s.notifyCriteria == nil {
		return false
	}
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	return s.notifyCriteria(s.buffer)
}

func (s *stepMonitor[I]) sendNotificationAndClearBuffer() {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()
	s.notify(s.buffer)
	s.buffer = make([]I, 0)
}

func (s *stepMonitor[I]) Run(ctx context.Context, wg *sync.WaitGroup) {
	if s.checkInterval == 0 {
		// 10 hours is to cover the case where the time is not set. The threshould should be met far before this time and the timer is reset.
		// this is needed to keep the thread alive as well inc case there is a long delay in the input.
		s.checkInterval = 100 * time.Hour
	}
	timer := time.NewTimer(s.checkInterval)
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
			if s.shouldNotify() {
				s.sendNotificationAndClearBuffer()
				timer.Reset(s.checkInterval)
			}
			// passing through the element to the next step
			s.output <- i
		case <-timer.C:
			s.sendNotificationAndClearBuffer()
			timer.Reset(s.checkInterval)
		}
	}
}
