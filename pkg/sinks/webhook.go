package sinks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type WebhookConfig struct {
	Endpoint string                 `yaml:"endpoint"`
	TLS      TLS                    `yaml:"tls"`
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
	tlsClientConfig, err := setupTLS(&w.cfg.TLS)
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}
	client := http.DefaultClient
	client.Transport = &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsClientConfig,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return errors.New("not successfull (2xx) response: " + string(body))
	}

	return nil
}
