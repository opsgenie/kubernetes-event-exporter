package sinks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type Zulip struct {
	cfg *ZulipConfig
}

type ZulipConfig struct {
	Endpoint string                 `yaml:"endpoint"`
	TLS      TLS                    `yaml:"tls"`
	Layout   map[string]interface{} `yaml:"layout"`
	Username string                 `yaml:"username"`
	Password string                 `yaml:"password"`
	Type     string                 `yaml:"type"`
	To       string                 `yaml:"to"`
	Topic    string                 `yaml:"topic"`
}

func NewZulip(cfg *ZulipConfig) (Sink, error) {
	return &Zulip{cfg: cfg}, nil
}

func (w *Zulip) Close() {
	// No-op
}

func (w *Zulip) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	event, err := serializeEventWithLayout(w.cfg.Layout, ev)
	if err != nil {
		return err
	}

	var eventData map[string]interface{}
	err = json.Unmarshal(event, &eventData)
	if err != nil {
		return err
	}
	involvedObject, err := json.MarshalIndent(eventData["involvedObject"], "", "  ")
	if err != nil {
		return err
	}

	output := fmt.Sprintf("# %s\n%s: %s\n```%s```",
		eventData["reason"],
		eventData["type"],
		eventData["message"],
		string(involvedObject),
	)

	data := url.Values{
		"type":    {w.cfg.Type},
		"to":      {w.cfg.To},
		"topic":   {w.cfg.Topic},
		"content": {output},
	}

	req, err := http.NewRequest(http.MethodPost, w.cfg.Endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(w.cfg.Username, w.cfg.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
