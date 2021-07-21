package sinks

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type FirehoseConfig struct {
	DeliveryStreamName string                 `yaml:"deliveryStreamName"`
	Region             string                 `yaml:"region"`
	Layout             map[string]interface{} `yaml:"layout"`
}

type FirehoseSink struct {
	cfg *FirehoseConfig
	svc *firehose.Firehose
}

func NewFirehoseSink(cfg *FirehoseConfig) (Sink, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return nil, err
	}

	return &FirehoseSink{
		cfg: cfg,
		svc: firehose.New(sess),
	}, nil
}

func (f *FirehoseSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	var toSend []byte

	if f.cfg.Layout != nil {
		res, err := convertLayoutTemplate(f.cfg.Layout, ev)
		if err != nil {
			return err
		}

		toSend, err = json.Marshal(res)
		if err != nil {
			return err
		}
	} else {
		toSend = ev.ToJSON()
	}

	_, err := f.svc.PutRecord(&firehose.PutRecordInput{
		Record: &firehose.Record{
			Data: toSend,
		},
		DeliveryStreamName: aws.String(f.cfg.DeliveryStreamName),
	})

	return err
}

func (f *FirehoseSink) Close() {
	// No-op
}
