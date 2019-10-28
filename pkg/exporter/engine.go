package exporter

import "github.com/opsgenie/kubernetes-event-exporter/pkg/kube"

// Engine is responsible for initializing the receivers from sinks
type Engine struct {
	Route    Route
	Registry ReceiverRegistry
}

func NewEngine(config *Config, registry ReceiverRegistry) *Engine {
	for _, v := range config.Receivers {

		registry.Register(v.Name, v.GetSink())
	}

	return &Engine{
		Route:    config.Route,
		Registry: registry,
	}
}

// OnEvent does not care whether event is add or update. Prior filtering should be done in the controller/watcher
func (e *Engine) OnEvent(event *kube.EnhancedEvent) {
	e.Route.ProcessEvent(event, e.Registry)
}
