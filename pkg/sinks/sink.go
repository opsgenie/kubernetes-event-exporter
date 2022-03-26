package sinks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
)

// Sink is the interface that the third-party providers should implement. It should just get the event and
// transform it depending on its configuration and submit it. Error handling for retries etc. should be handled inside
// for now.
type Sink interface {
	Send(ctx context.Context, ev *kube.EnhancedEvent) error
	Close()
}

// BatchSink is an extension Sink that can handle batch events.
// NOTE: Currently no provider implements it nor the receivers can handle it.
type BatchSink interface {
	Sink
	SendBatch([]*kube.EnhancedEvent) error
}

type TLS struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	ServerName         string `yaml:"serverName"`
	CaFile             string `yaml:"caFile"`
	KeyFile            string `yaml:"keyFile"`
	CertFile           string `yaml:"certFile"`
}

func setupTLS(cfg *TLS) (*tls.Config, error) {
	var caCert []byte

	if len(cfg.CaFile) > 0 {
		readFile, err := ioutil.ReadFile(cfg.CaFile)
		if err != nil {
			return nil, err
		}
		caCert = readFile
	}

	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.ServerName,
	}
	if len(cfg.KeyFile) > 0 && len(cfg.CertFile) > 0 {
		tlsClientConfig.RootCAs = x509.NewCertPool()
		tlsClientConfig.RootCAs.AppendCertsFromPEM(caCert)

		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not read client certificate or key: %w", err)
		}
		tlsClientConfig.Certificates = append(tlsClientConfig.Certificates, cert)
	}
	if len(cfg.KeyFile) > 0 && len(cfg.CertFile) == 0 {
		return nil, errors.New("configured keyFile but forget certFile for client certificate authentication")
	}
	if len(cfg.KeyFile) == 0 && len(cfg.CertFile) > 0 {
		return nil, errors.New("configured certFile but forget keyFile for client certificate authentication")
	}
	return tlsClientConfig, nil
}
