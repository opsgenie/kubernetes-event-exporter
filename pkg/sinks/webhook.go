package sinks

import (
	"bytes"
	"errors"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"net/http"
)

type WebhookConfig struct {
	Endpoint string
}

func (w *WebhookConfig) Validate() error {
	return nil
}

type Webhook struct {
	Config *WebhookConfig
}

func (w *Webhook) Send(ev *kube.EnhancedEvent) error {
	resp, err := http.Post(w.Config.Endpoint, "application/json", bytes.NewReader(ev.ToJSON()))
	if err != nil {
		return nil
	}

	// TODO: make this pretty please
	if resp.StatusCode != http.StatusOK {
		return errors.New("not 200")
	}

	return nil
}
