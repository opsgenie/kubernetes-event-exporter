package sinks

import (
	"bytes"
	"context"
	"errors"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"io/ioutil"
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

func (w *Webhook) Close() {
	// No-op
}

func (w *Webhook) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	resp, err := http.Post(
		w.Config.Endpoint,
		"application/json",
		bytes.NewReader(ev.ToJSON()),
	)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// TODO: make this prettier please
	if resp.StatusCode != http.StatusOK {
		return errors.New("not 200: " + string(body))
	}

	return nil
}
