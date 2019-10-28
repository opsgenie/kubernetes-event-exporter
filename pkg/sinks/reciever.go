package sinks

// Receiver allows receiving
type ReceiverConfig struct {
	Name     string
	Webhook  *WebhookConfig
	InMemory *InMemoryConfig
}

func (r *ReceiverConfig) Validate() error {
	return nil
}

func (r *ReceiverConfig) GetSink() Sink {
	// Sorry for this code, but its Go
	if r.Webhook != nil {
		return &Webhook{Config: r.Webhook}
	}
	if r.InMemory != nil {
		// This reference is used for test purposes to count the events in the sink.
		// It should not be used in production since it will only cause memory leak and (b)OOM
		sink := &InMemory{Config: r.InMemory}
		r.InMemory.Ref = sink
		return sink
	}

	panic("invalid sink: " + r.Name)
}
