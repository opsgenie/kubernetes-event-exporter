package exporter

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/sinks"
)

// SyncRegistry is for development purposes and performs poorly and blocks when an event is received so it is
// not suited for high volume & production workloads
type SyncRegistry struct {
	reg map[string]sinks.Sink
}

func (s *SyncRegistry) SendEvent(name string, event *kube.EnhancedEvent) {
	_ = s.reg[name].Send(event)
}

func (s *SyncRegistry) Register(name string, sink sinks.Sink) {
	if s.reg == nil {
		s.reg = make(map[string]sinks.Sink)
	}

	s.reg[name] = sink
}
