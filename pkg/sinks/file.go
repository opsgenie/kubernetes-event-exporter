package sinks

import (
	"context"
	"encoding/json"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"os"
)

type FileConfig struct {
	Path   string                 `yaml:"file"`
	Layout map[string]interface{} `yaml:"layout"`
}

func (f *FileConfig) Validate() error {
	return nil
}

type File struct {
	file    *os.File
	encoder *json.Encoder
	layout  map[string]interface{}
}

func NewFileSink(config *FileConfig) (*File, error) {
	file, err := os.OpenFile(config.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	return &File{
		file:    file,
		encoder: json.NewEncoder(file),
		layout:  config.Layout,
	}, nil
}

func (f *File) Close() {
	_ = f.file.Close()
}

func (f *File) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	if f.layout == nil {
		return f.encoder.Encode(ev)
	}

	res, err := convertLayoutTemplate(f.layout, ev)
	if err != nil {
		return err
	}

	return f.encoder.Encode(res)
}
