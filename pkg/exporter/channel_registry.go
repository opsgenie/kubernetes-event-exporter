package exporter

import (
	"context"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/sinks"
	"github.com/rs/zerolog/log"
)

// ChannelBasedReceiverRegistry creates two channels for each receiver. One is for receiving events and other one is
// for breaking out of the infinite loop. Each message is passed to receivers
// This might not be the best way to implement such feature. A ring buffer can be better
// and we might need a mechanism to drop the vents
type ChannelBasedReceiverRegistry struct {
	ch     map[string]chan kube.EnhancedEvent
	exitCh map[string]chan interface{}
}

func (r *ChannelBasedReceiverRegistry) SendEvent(name string, event *kube.EnhancedEvent) {
	ch := r.ch[name]
	if ch == nil {
		log.Error().Str("name", name).Msg("There is no channel")
	}

	go func() {
		ch <- *event
	}()
}

func (r *ChannelBasedReceiverRegistry) Register(name string, receiver sinks.Sink) {
	if r.ch == nil {
		r.ch = make(map[string]chan kube.EnhancedEvent)
		r.exitCh = make(map[string]chan interface{})
	}

	ch := make(chan kube.EnhancedEvent)
	exitCh := make(chan interface{})

	r.ch[name] = ch
	r.exitCh[name] = exitCh

	go func() {
		for {
			select {
			case ev := <-ch:
				log.Debug().Str("sink", name).Str("event", ev.Message).Msg("sending event to sink")
				err := receiver.Send(context.Background(), &ev)
				if err != nil {
					log.Debug().Err(err).Str("sink", name).Str("event", ev.Message).Msg("Cannot send event")
				}
			case <-exitCh:
				log.Info().Str("receiver", name).Msg("Killing the receiver")
				break
			}
		}
	}()
}
