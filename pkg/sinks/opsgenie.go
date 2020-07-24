package sinks

import (
	"context"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type OpsgenieConfig struct {
	ApiKey      string            `yaml:"apiKey"`
	URL         client.ApiUrl     `yaml:"URL"`
	Priority    string            `yaml:"priority"`
	Message     string            `yaml:"message"`
	Alias       string            `yaml:"alias"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Details     map[string]string `yaml:"details"`
}

type OpsgenieSink struct {
	cfg         *OpsgenieConfig
	alertClient *alert.Client
}

func NewOpsgenieSink(config *OpsgenieConfig) (Sink, error) {
	if config.URL == "" {
		config.URL = client.API_URL
	}

	if config.Priority == "" {
		config.Priority = "P3"
	}

	alertClient, err := alert.NewClient(&client.Config{
		ApiKey:         config.ApiKey,
		OpsGenieAPIURL: config.URL,
	})

	if err != nil {
		return nil, err
	}

	return &OpsgenieSink{
		cfg:         config,
		alertClient: alertClient,
	}, nil
}

func (o *OpsgenieSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	request := alert.CreateAlertRequest{
		Priority: alert.Priority(o.cfg.Priority),
	}

	msg, err := GetString(ev, o.cfg.Message)
	if err != nil {
		return err
	}
	request.Message = msg

	// Alias is optional although highly recommended to work
	if o.cfg.Alias != "" {
		alias, err := GetString(ev, o.cfg.Alias)
		if err != nil {
			return err
		}
		request.Alias = alias
	}

	description, err := GetString(ev, o.cfg.Description)
	if err != nil {
		return err
	}
	request.Description = description

	if o.cfg.Tags != nil {
		tags := make([]string, 0)
		for _, v := range o.cfg.Tags {
			tag, err := GetString(ev, v)
			if err != nil {
				return err
			}
			tags = append(tags, tag)
		}
		request.Tags = tags
	}

	if o.cfg.Details != nil {
		details := make(map[string]string)
		for k, v := range o.cfg.Details {
			detail, err := GetString(ev, v)
			if err != nil {
				return err
			}
			details[k] = detail
		}
		request.Details = details
	}

	_, err = o.alertClient.Create(ctx, &request)
	return err
}

func (o *OpsgenieSink) Close() {
	// No-op
}
