package sinks

import "errors"

// Receiver allows receiving
type ReceiverConfig struct {
	Name          string
	Webhook       *WebhookConfig
	InMemory      *InMemoryConfig
	File          *FileConfig
	Elasticsearch *ElasticsearchConfig
	Kinesis       *KinesisConfig
	Opsgenie      *OpsgenieConfig
	SQS           *SQSConfig
	SNS           *SNSConfig
	Slack         *SlackConfig
}

func (r *ReceiverConfig) Validate() error {
	return nil
}

func (r *ReceiverConfig) GetSink() (Sink, error) {
	if r.InMemory != nil {
		// This reference is used for test purposes to count the events in the sink.
		// It should not be used in production since it will only cause memory leak and (b)OOM
		sink := &InMemory{Config: r.InMemory}
		r.InMemory.Ref = sink
		return sink, nil
	}

	// Sorry for this code, but its Go
	if r.Webhook != nil {
		return NewWebhook(r.Webhook)
	}

	if r.File != nil {
		return NewFileSink(r.File)
	}

	if r.Elasticsearch != nil {
		return NewElasticsearch(r.Elasticsearch)
	}

	if r.Kinesis != nil {
		return NewKinesisSink(r.Kinesis)
	}

	if r.Opsgenie != nil {
		return NewOpsgenieSink(r.Opsgenie)
	}

	if r.SQS != nil {
		return NewSQSSink(r.SQS)
	}

	if r.SNS != nil {
		return NewSNSSink(r.SNS)
	}

	if r.Slack != nil {
		return NewSlackSink(r.Slack)
	}

	return nil, errors.New("unknown sink")
}
