package sinks

import (
	"context"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

// Sink is the interface that the third-party providers should implement. It should just get the event and
// transform it depending on its configuration and submit it. Error handling for retries etc. should be handled inside
// for now.
type Sink interface {
	Send(ctx context.Context, ev *kube.EnhancedEvent) error
	Close()
}

// BatchSink is an extension Sink that can handle batch events.
// NOTE: Currently no provider implements it nor the receivers can handle it.
type BatchSink interface {
	Sink
	SendBatch([]*kube.EnhancedEvent) error
}
