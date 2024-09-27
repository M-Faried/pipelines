package pipelines

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockNotify struct {
	called  bool
	counter int
}

func (m *mockNotify) Handle() {
	m.called = true
	m.counter++
}

func TestShouldNotify(t *testing.T) {

	tests := []struct {
		name           string
		buffer         []int
		notifyCriteria StepMonitorNotifyCriteria[int]
		expected       bool
	}{
		{
			name:           "NotifyCriteria is nil",
			buffer:         []int{1, 2, 3},
			notifyCriteria: nil,
			expected:       false,
		},
		{
			name:   "Buffer meets criteria",
			buffer: []int{1, 2, 3},
			notifyCriteria: func(buffer []int) bool {
				return len(buffer) >= 3
			},
			expected: true,
		},
		{
			name:   "Buffer does not meet criteria",
			buffer: []int{1, 2},
			notifyCriteria: func(buffer []int) bool {
				return len(buffer) >= 3
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepMonitor[int]{
				notifyCriteria: tt.notifyCriteria,
				buffer:         tt.buffer,
				checkInterval:  0,
			}
			if got := s.shouldNotify(); got != tt.expected {
				t.Errorf("shouldNotify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepMonitorNotifyPeriodicCall(t *testing.T) {

	mockNotify := &mockNotify{}
	s := &stepMonitor[int]{
		stepBase: stepBase[int]{
			label:  "Testing Intrval Check",
			input:  make(chan int),
			output: make(chan int),
		},
		notifyCriteria: nil,
		notify:         mockNotify.Handle,
		checkInterval:  100 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go s.Run(ctx, &wg)
	time.Sleep(400 * time.Millisecond)

	close(s.input)
	close(s.output)

	wg.Wait()

	if !mockNotify.called {
		t.Errorf("expected notify to be called")
	}
	if mockNotify.counter < 3 {
		t.Errorf("expected notify to be called 3 times, got %v", mockNotify.counter)
	}
}
