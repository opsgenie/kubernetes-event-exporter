package sinks

import (
	"context"
	"github.com/nlopes/slack"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type SlackConfig struct {
	Token   string
	Channel string
	Message string
	Fields  map[string]string
}

type SlackSink struct {
	cfg    *SlackConfig
	client *slack.Client
}

func NewSlackSink(cfg *SlackConfig) (Sink, error) {
	return &SlackSink{
		cfg:    cfg,
		client: slack.New(cfg.Token),
	}, nil
}

func (s *SlackSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	channel, err := GetString(ev, s.cfg.Channel)
	if err != nil {
		return err
	}

	message, err := GetString(ev, s.cfg.Message)
	if err != nil {
		return err
	}

	options := []slack.MsgOption{slack.MsgOptionText(message, true)}
	if s.cfg.Fields != nil {
		fields := make([]slack.AttachmentField, 0)
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
		options = append(options, slack.MsgOptionAttachments(slack.Attachment{Fields: fields}))
	}

	_, _, _, err = s.client.SendMessageContext(ctx, channel, options...)
	return err
}

func (s *SlackSink) Close() {
	// No-op
}
