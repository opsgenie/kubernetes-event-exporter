package sinks

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

type ElasticsearchConfig struct {
	// Connection specific
	Hosts    []string `yaml:"hosts"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	CloudID  string   `yaml:"cloudID"`
	APIKey   string   `yaml:"apiKey"`
	// Indexing preferences
	UseEventID bool `yaml:"useEventID"`
	// DeDot all labels and annotations in the event. For both the event and the involvedObject
	DeDot       bool   `yaml:"deDot"`
	Index       string `yaml:"index"`
	IndexFormat string `yaml:"indexFormat"`
	Type        string `yaml:"type"`
	TLS         struct {
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
		ServerName         string `yaml:"serverName"`
		CaFile             string `yaml:"caFile"`
		KeyFile            string `yaml:"keyFile"`
		CertFile           string `yaml:"certFile"`
	} `yaml:"tls"`
	Layout map[string]interface{} `yaml:"layout"`
}

func NewElasticsearch(cfg *ElasticsearchConfig) (*Elasticsearch, error) {

	tlsClientConfig, err := setupTLS(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup TLS: %w", err)
	}

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

func setupTLS(cfg *ElasticsearchConfig) (*tls.Config, error) {
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
	if len(cfg.TLS.KeyFile) > 0 && len(cfg.TLS.CertFile) > 0 {
		tlsClientConfig.RootCAs = x509.NewCertPool()
		tlsClientConfig.RootCAs.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not read client certificate or key: %w", err)
		}
		tlsClientConfig.Certificates = append(tlsClientConfig.Certificates, cert)
	}
	if len(cfg.TLS.KeyFile) > 0 && len(cfg.TLS.CertFile) == 0 {
		return nil, errors.New("configured keyFile but forget certFile for client certificate authentication")
	}
	if len(cfg.TLS.KeyFile) == 0 && len(cfg.TLS.CertFile) > 0 {
		return nil, errors.New("configured certFile but forget keyFile for client certificate authentication")
	}
	return tlsClientConfig, nil
}

type Elasticsearch struct {
	client *elasticsearch.Client
	cfg    *ElasticsearchConfig
}

var regex = regexp.MustCompile(`(?s){(.*)}`)

func formatIndexName(pattern string, when time.Time) string {
	m := regex.FindAllStringSubmatchIndex(pattern, -1)
	current := 0
	var builder strings.Builder

	for i := 0; i < len(m); i++ {
		pair := m[i]

		builder.WriteString(pattern[current:pair[0]])
		builder.WriteString(when.Format(pattern[pair[0]+1 : pair[1]-1]))
		current = pair[1]
	}

	builder.WriteString(pattern[current:])

	return builder.String()
}

func (e *Elasticsearch) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	var toSend []byte

	if e.cfg.DeDot {
		de := ev.DeDot()
		ev = &de
	}
	if e.cfg.Layout != nil {
		res, err := convertLayoutTemplate(e.cfg.Layout, ev)
		if err != nil {
			return err
		}

		toSend, err = json.Marshal(res)
		if err != nil {
			return err
		}
	} else {
		toSend = ev.ToJSON()
	}

	var index string
	if len(e.cfg.IndexFormat) > 0 {
		now := time.Now()
		index = formatIndexName(e.cfg.IndexFormat, now)
	} else {
		index = e.cfg.Index
	}

	req := esapi.IndexRequest{
		Body:  bytes.NewBuffer(toSend),
		Index: index,
	}

	// This should not be used for clusters with ES8.0+.
	if len(e.cfg.Type) > 0 {
		req.DocumentType = e.cfg.Type
	}

	if e.cfg.UseEventID {
		req.DocumentID = string(ev.UID)
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		rb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Error().Msgf("Indexing failed: %s", string(rb))
	}
	return nil
}

func (e *Elasticsearch) Close() {
	// No-op
}
