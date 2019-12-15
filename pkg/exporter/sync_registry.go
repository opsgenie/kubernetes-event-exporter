package exporter

import (
	"context"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/sinks"
	"github.com/rs/zerolog/log"
)

// SyncRegistry is for development purposes and performs poorly and blocks when an event is received so it is
// not suited for high volume & production workloads
type SyncRegistry struct {
	reg map[string]sinks.Sink
}

func (s *SyncRegistry) SendEvent(name string, event *kube.EnhancedEvent) {
	err := s.reg[name].Send(context.Background(), event)
	if err != nil {
		log.Debug().Err(err).Str("sink", name).Str("event", string(event.UID)).Msg("Cannot send event")
	}
}

func (s *SyncRegistry) Register(name string, sink sinks.Sink) {
	if s.reg == nil {
		s.reg = make(map[string]sinks.Sink)
	}

	s.reg[name] = sink
}

func (s *SyncRegistry) Close() {
	for name, sink := range s.reg {
		log.Info().Str("sink", name).Msg("Closing sink")
		sink.Close()
	}
}
