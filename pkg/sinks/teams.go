package sinks

import (
	"bytes"
	"fmt"
	"encoding/json"
	"context"
	"errors"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"io/ioutil"
	"net/http"
)

type TeamsConfig struct {
	Endpoint string                 `yaml:"endpoint"`
	Layout   map[string]interface{} `yaml:"layout"`
	Headers  map[string]string      `yaml:"headers"`
}

func NewTeamsSink(cfg *TeamsConfig) (Sink, error) {
	return &Teams{cfg: cfg}, nil
}

type Teams struct {
	cfg *TeamsConfig
}

func (w *Teams) Close() {
	// No-op
}

func (w *Teams) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	event, err := serializeEventWithLayout(w.cfg.Layout, ev)
	if err != nil {
		return err
	}
	
	var eventData map[string]interface{}
	json.Unmarshal([]byte(event), &eventData)
	output := fmt.Sprintf("Event: %s \nStatus: %s \nMetadata: %s", eventData["message"],  eventData["reason"], eventData["metadata"])

	reqBody, err := json.Marshal(map[string]string{
		"summary": "event",
		"text": string([]byte(output)),
	 })

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
