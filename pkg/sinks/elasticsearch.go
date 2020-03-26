package sinks

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"io/ioutil"
	"net/http"
)

type ElasticsearchConfig struct {
	// Connection specific
	Hosts    []string `yaml:"hosts"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	CloudID  string   `yaml:"cloudID"`
	APIKey   string   `yaml:"apiKey"`
	// Indexing preferences
	UseEventID bool   `yaml:"useEventID"`
	Index      string `yaml:"index"`
	TLS        struct {
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
		ServerName         string `yaml:"serverName"`
		CaFile             string `yaml:"caFile"`
	} `yaml:"tls"`
}

func NewElasticsearch(cfg *ElasticsearchConfig) (*Elasticsearch, error) {
	var caCert []byte

	if len(cfg.TLS.CaFile) > 0 {
		readFile, err := ioutil.ReadFile(cfg.TLS.CaFile)
		if err != nil {
			return nil, err
		}
		caCert = readFile
	}

	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		ServerName:         cfg.TLS.ServerName,
	}
	tlsClientConfig.RootCAs.AppendCertsFromPEM(caCert)

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Hosts,
		Username:  cfg.Username,
		Password:  cfg.Password,
		CloudID:   cfg.CloudID,
		APIKey:    cfg.APIKey,
		Transport: &http.Transport{
			TLSClientConfig: tlsClientConfig,
		},
	})
	if err != nil {
		return nil, err
	}

	return &Elasticsearch{
		client: client,
		cfg:    cfg,
	}, nil
}

type Elasticsearch struct {
	client *elasticsearch.Client
	cfg    *ElasticsearchConfig
}

func (e *Elasticsearch) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Body:  bytes.NewBuffer(b),
		Index: e.cfg.Index,
	}

	if e.cfg.UseEventID {
		req.DocumentID = string(ev.UID)
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	_ = resp.Body
	return nil
}

func (e *Elasticsearch) Close() {
	// No-op
}
