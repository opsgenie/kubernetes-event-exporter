package sinks

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

type PipeConfig struct {
	Path       string                 `yaml:"path"`
	Layout     map[string]interface{} `yaml:"layout"`
}

func (f *PipeConfig) Validate() error {
	return nil
}

type Pipe struct {
	writer  io.WriteCloser
	encoder *json.Encoder
	layout  map[string]interface{}
}

func NewPipeSink(config *PipeConfig) (*Pipe, error) {
	mode := os.FileMode(0644)
	f, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
    if err != nil {
        return nil,err
    }
	return &Pipe{
		writer:  f,
		encoder: json.NewEncoder(f),
		layout:  config.Layout,
	}, nil
}

func (f *Pipe) Close() {
	_ = f.writer.Close()
}

func (f *Pipe) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	if f.layout == nil {
		return f.encoder.Encode(ev)
	}

	res, err := convertLayoutTemplate(f.layout, ev)
	if err != nil {
		return err
	}

	return f.encoder.Encode(res)
}
