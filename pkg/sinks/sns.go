package sinks

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type SNSConfig struct {
	TopicARN string                 `yaml:"topicARN"`
	Region   string                 `yaml:"region"`
	Layout   map[string]interface{} `yaml:"layout"`
}

type SNSSink struct {
	cfg *SNSConfig
	svc *sns.SNS
}

func NewSNSSink(cfg *SNSConfig) (Sink, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region)},
	)
	if err != nil {
		return nil, err
	}

	svc := sns.New(sess)
	return &SNSSink{
		cfg: cfg,
		svc: svc,
	}, nil
}

func (s *SNSSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	toSend, e := serializeEventWithLayout(s.cfg.Layout, ev)
	if e != nil {
		return e
	}

	_, err := s.svc.PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(string(toSend)),
		TopicArn: aws.String(s.cfg.TopicARN),
	})

	return err
}

func (s *SNSSink) Close() {
}
