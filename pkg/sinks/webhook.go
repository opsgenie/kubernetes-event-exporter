package sinks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type WebhookConfig struct {
	Endpoint string                 `yaml:"endpoint"`
	TLS      TLS                    `yaml:"tls"`
	Layout   map[string]interface{} `yaml:"layout"`
	Headers  map[string]string      `yaml:"headers"`
}

type serializeEvent func(layout map[string]interface{}, ev *kube.EnhancedEvent) ([]byte, error)

var mimetypeConverter = map[string]serializeEvent{
	"application/json": serializeEventWithLayout,
	"application/xml":  serializeXMLEventWithLayout,
}

func NewWebhook(cfg *WebhookConfig) (Sink, error) {

	var contentType = ""
	var converter serializeEvent

	for k, v := range cfg.Headers {

		if strings.Contains(strings.ToLower(k), "content-type") {
			contentType = v
			break
		}
	}

	if contentType == "" {
		// default
		contentType = "application/json"
		cfg.Headers["Content-Type"] = "application/json"
	}

	for k, v := range mimetypeConverter {
		if strings.Contains(contentType, k) {
			converter = v
			break
		}
	}

	if converter == nil {

		errMsg := "Wrong Content-Type in headers: only %s are supported"
		mimetypes := ""
		for k := range mimetypeConverter {
			mimetypes = mimetypes + ", " + k
		}

		return nil, fmt.Errorf(errMsg, mimetypes)
	}

	return &Webhook{cfg: cfg, serializer: converter}, nil
}

type Webhook struct {
	cfg        *WebhookConfig
	serializer serializeEvent
}

func (w *Webhook) Close() {
	// No-op
}

func (w *Webhook) Send(ctx context.Context, ev *kube.EnhancedEvent) error {

	reqBody, err := w.serializer(w.cfg.Layout, ev)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, w.cfg.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	for k, v := range w.cfg.Headers {
		req.Header.Add(k, v)
	}

	tlsClientConfig, err := setupTLS(&w.cfg.TLS)
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}
	client := http.DefaultClient
	client.Transport = &http.Transport{
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
