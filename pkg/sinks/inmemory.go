package sinks

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type InMemoryConfig struct {
	Ref *InMemory
}

type InMemory struct {
	Events []*kube.EnhancedEvent
	Config *InMemoryConfig
}

func (i *InMemory) Send(ev *kube.EnhancedEvent) error {
	i.Events = append(i.Events, ev)
	return nil
}
