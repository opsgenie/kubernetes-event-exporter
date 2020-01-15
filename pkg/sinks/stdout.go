package sinks

import (
	"context"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

// StdoutConfig is config of StdoutSink
type StdoutConfig struct {
	Layout map[string]interface{} `yaml:"layout"`
}

// NewStdoutSink return new stdout sink
func NewStdoutSink(cfg *StdoutConfig) (*Stdout, error) {
	return &Stdout{
		cfg: cfg,
	}, nil
}

// Stdout print event to stdout
type Stdout struct {
	cfg *StdoutConfig
}

// Send ...
func (s *Stdout) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	if s.cfg.Layout != nil {
		res, err := convertLayoutTemplate(s.cfg.Layout, ev)
		if err != nil {
			return err
		}
		_ = res
		log.Info().Fields(res).Msg("")
	} else {
		log.Info().RawJSON("event", ev.ToJSON()).Msg("")
	}
	return nil
}

// Close ...
func (s *Stdout) Close() {
	// pass
}
