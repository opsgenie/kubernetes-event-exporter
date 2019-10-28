package opsgenie

import (
	"github.com/opsgenie/kubernetes-event-exporter/pkg/config"
	"net/http"
)

type Notifier struct {
	cfg    *config.OpsgenieConfig
	client http.Client
}

func New(cfg *config.OpsgenieConfig) *Notifier {
	return &Notifier{
		cfg:    cfg,
		client: http.Client{},
	}
}
