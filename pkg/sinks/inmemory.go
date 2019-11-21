package sinks

import (
	"context"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type InMemoryConfig struct {
	Ref *InMemory
}

type InMemory struct {
	Events []*kube.EnhancedEvent
	Config *InMemoryConfig
}

func (i *InMemory) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	i.Events = append(i.Events, ev)
	return nil
}

func (i *InMemory) Close() {
	// No-op
}


