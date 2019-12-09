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
	Endpoint string                 `yaml:"endpoint"`
	Layout   map[string]interface{} `yaml:"layout"`
	Headers  map[string]string      `yaml:"headers"`
}

func NewWebhook(cfg *WebhookConfig) (Sink, error) {
	return &Webhook{cfg: cfg}, nil
}

type Webhook struct {
	cfg *WebhookConfig
}

func (w *Webhook) Close() {
	// No-op
}

func (w *Webhook) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	reqBody, err := serializeEventWithLayout(w.cfg.Layout, ev)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, w.cfg.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	for k, v := range w.cfg.Headers {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
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
