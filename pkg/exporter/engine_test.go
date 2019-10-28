package exporter

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/sinks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEngineNoRoutes(t *testing.T) {
	cfg := &Config{
		Route:     Route{},
		Receivers: nil,
	}

	e := NewEngine(cfg, &SyncRegistry{})
	ev := &kube.EnhancedEvent{}
	e.OnEvent(ev)
}

func TestEngineSimple(t *testing.T) {
	config := &sinks.InMemoryConfig{}
	cfg := &Config{
		Route: Route{
			Match: []Rule{{
				Receiver: "in-mem",
			}},
		},
		Receivers: []sinks.ReceiverConfig{{
			Name:     "in-mem",
			InMemory: config,
		}},
	}

	e := NewEngine(cfg, &SyncRegistry{})
	ev := &kube.EnhancedEvent{}
	e.OnEvent(ev)

	assert.Contains(t, config.Ref.Events, ev)
}

func TestEngineDropSimple(t *testing.T) {
	config := &sinks.InMemoryConfig{}
	cfg := &Config{
		Route: Route{
			Drop: []Rule{{
				// Drops anything
			}},
			Match: []Rule{{
				Receiver: "in-mem",
			}},
		},
		Receivers: []sinks.ReceiverConfig{{
			Name:     "in-mem",
			InMemory: config,
		}},
	}

	e := NewEngine(cfg, &SyncRegistry{})
	ev := &kube.EnhancedEvent{}
	e.OnEvent(ev)

	assert.NotContains(t, config.Ref.Events, ev)
	assert.Empty(t, config.Ref.Events)
}
