package sinks

import (
	"context"

	"github.com/nlopes/slack"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type SlackConfig struct {
	Endpoint string            `yaml:endpoint`
	Message  string            `yaml:"message"`
	Fields   map[string]string `yaml:"fields"`
}

type SlackSink struct {
	cfg *SlackConfig
}

func NewSlackSink(cfg *SlackConfig) (Sink, error) {
	return &SlackSink{
		cfg: cfg,
	}, nil
}

func (s *SlackSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	message, err := GetString(ev, s.cfg.Message)
	if err != nil {
		return err
	}

	endpoint, err := GetString(ev, s.cfg.Endpoint)
	if err != nil {
		return err
	}

	fields := make([]slack.AttachmentField, 0)

	if s.cfg.Fields != nil {
		for k, v := range s.cfg.Fields {
			fieldText, err := GetString(ev, v)
			if err != nil {
				return err
			}

			fields = append(fields, slack.AttachmentField{
				Title: k,
				Value: fieldText,
				Short: false,
			})
		}
	}

	attachment := slack.Attachment{
		Fields: fields,
	}

	webhookMessage := slack.WebhookMessage{
		Text:        message,
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(endpoint, &webhookMessage)
	if err != nil {
		return err
	}
	return nil
}

func (s *SlackSink) Close() {
	// No-op
}
