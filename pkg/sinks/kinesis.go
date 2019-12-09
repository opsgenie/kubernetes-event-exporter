package sinks

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type KinesisConfig struct {
	StreamName string                 `yaml:"streamName"`
	Region     string                 `yaml:"region"`
	Layout     map[string]interface{} `yaml:"layout"`
}

type KinesisSink struct {
	cfg *KinesisConfig
	svc *kinesis.Kinesis
}

func NewKinesisSink(cfg *KinesisConfig) (Sink, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return nil, err
	}

	return &KinesisSink{
		cfg: cfg,
		svc: kinesis.New(sess),
	}, nil
}

func (k *KinesisSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	var toSend []byte

	if k.cfg.Layout != nil {
		res, err := convertLayoutTemplate(k.cfg.Layout, ev)
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

	_, err := k.svc.PutRecord(&kinesis.PutRecordInput{
		Data:         toSend,
		PartitionKey: aws.String(string(ev.UID)),
		StreamName:   aws.String(k.cfg.StreamName),
	})

	return err
}

func (k *KinesisSink) Close() {
	// No-op
}
