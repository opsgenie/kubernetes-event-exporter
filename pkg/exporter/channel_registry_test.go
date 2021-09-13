package exporter

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/stretchr/testify/assert"
)

type SlowSink struct {
	wg   *sync.WaitGroup
	rcvd []*kube.EnhancedEvent
}

func NewSlowSink() *SlowSink {
	var rcvd []*kube.EnhancedEvent
	return &SlowSink{
		rcvd: rcvd,
		wg:   &sync.WaitGroup{},
	}
}

func (s *SlowSink) Send(ctx context.Context, event *kube.EnhancedEvent) error {
	s.wg.Add(1)
	time.Sleep(2 * time.Second)
	s.rcvd = append(s.rcvd, event)
	s.wg.Done()
	return nil
}

func (s *SlowSink) Close() {
	s.wg.Wait()
}

func AlmostEqualDuration(a time.Duration, b time.Duration, equalityThreshold time.Duration) bool {
	diff := math.Abs(float64(a) - float64(b))
	return diff <= float64(equalityThreshold)
}

func TestSendEventAsync(t *testing.T) {
	name := "test"
	cr := &ChannelBasedReceiverRegistry{}
	sink := NewSlowSink()
	cr.Register(name, sink)
	ev := &kube.EnhancedEvent{}

	start := time.Now()
	cr.SendEvent(name, ev)
	cr.SendEvent(name, ev)
	cr.SendEvent(name, ev)
	// Sleep for a very short period to ensure that
	// all events have been sent to the sink channels
	time.Sleep(10 * time.Millisecond)
	cr.Close()
	elapsed := time.Since(start)
	assert.True(t, len(sink.rcvd) == 3)
	assert.True(t, AlmostEqualDuration(elapsed, 2*time.Second, 200*time.Millisecond))
}
