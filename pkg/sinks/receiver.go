package sinks

import "errors"

// Receiver allows receiving
type ReceiverConfig struct {
	Name          string               `yaml:"name"`
	InMemory      *InMemoryConfig      `yaml:"inMemory"`
	Webhook       *WebhookConfig       `yaml:"webhook"`
	File          *FileConfig          `yaml:"file"`
	Stdout        *StdoutConfig        `yaml:"stdout"`
	Elasticsearch *ElasticsearchConfig `yaml:"elasticsearch"`
	Kinesis       *KinesisConfig       `yaml:"kinesis"`
	Opsgenie      *OpsgenieConfig      `yaml:"opsgenie"`
	SQS           *SQSConfig           `yaml:"sqs"`
	SNS           *SNSConfig           `yaml:"sns"`
	Slack         *SlackConfig         `yaml:"slack"`
	Kafka         *KafkaConfig         `yaml:"kafka"`
	Pubsub        *PubsubConfig        `yaml:"pubsub"`
	Opscenter     *OpsCenterConfig     `yaml:"opscenter"`
	Teams         *TeamsConfig         `yaml:"teams"`
	BigQuery      *BigQueryConfig      `yaml:"bigquery"`
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

	if r.Stdout != nil {
		return NewStdoutSink(r.Stdout)
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

	if r.Kafka != nil {
		return NewKafkaSink(r.Kafka)
	}

	if r.Pubsub != nil {
		return NewPubsubSink(r.Pubsub)
	}

	if r.Opscenter != nil {
		return NewOpsCenterSink(r.Opscenter)
	}

	if r.Teams != nil {
		return NewTeamsSink(r.Teams)
	}

	if r.BigQuery != nil {
		return NewBigQuerySink(r.BigQuery)
	}

	return nil, errors.New("unknown sink")
}
