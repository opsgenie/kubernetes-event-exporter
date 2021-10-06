package sinks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"

	"github.com/Shopify/sarama"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

// KafkaConfig is the Kafka producer configuration
type KafkaConfig struct {
	Topic   string                 `yaml:"topic"`
	Brokers []string               `yaml:"brokers"`
	Layout  map[string]interface{} `yaml:"layout"`
	TLS     struct {
		Enable             bool   `yaml:"enable"`
		CaFile             string `yaml:"caFile"`
		CertFile           string `yaml:"certFile"`
		KeyFile            string `yaml:"keyFile"`
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	} `yaml:"tls"`
	KafkaEncode Avro `yaml:"avro"`
}

// KafkaEncoder is an interface type for adding an
// encoder to the kafka data pipeline
type KafkaEncoder interface {
	encode([]byte) ([]byte, error)
}

// KafkaSink is a sink that sends events to a Kafka topic
type KafkaSink struct {
	producer sarama.SyncProducer
	cfg      *KafkaConfig
	encoder  KafkaEncoder
}

func NewKafkaSink(cfg *KafkaConfig) (Sink, error) {
	var avro KafkaEncoder
	producer, err := createSaramaProducer(cfg)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("kafka: Producer initialized for topic: %s, brokers: %s", cfg.Topic, cfg.Brokers)
	if len(cfg.KafkaEncode.SchemaID) > 0 {
		var err error
		avro, err = NewAvroEncoder(cfg.KafkaEncode.SchemaID, cfg.KafkaEncode.Schema)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("kafka: Producer using avro encoding with schemaid: %s", cfg.KafkaEncode.SchemaID)
	}

	return &KafkaSink{
		producer: producer,
		cfg:      cfg,
		encoder:  avro,
	}, nil
}

// Send an event to Kafka synchronously
func (k *KafkaSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	var toSend []byte

	if k.cfg.Layout != nil {
		res, err := convertLayoutTemplate(k.cfg.Layout, ev)
		if err != nil {
			return err
		}

		toSend, err = json.Marshal(res)
		if err != nil {
			return err
		}
	} else if len(k.cfg.KafkaEncode.SchemaID) > 0 {
		var err error
		toSend, err = k.encoder.encode(ev.ToJSON())
		if err != nil {
			return err
		}
	} else {
		toSend = ev.ToJSON()
	}

	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: k.cfg.Topic,
		Key:   sarama.StringEncoder(string(ev.UID)),
		Value: sarama.ByteEncoder(toSend),
	})

	return err
}

// Close the Kafka producer
func (k *KafkaSink) Close() {
	log.Info().Msgf("kafka: Closing producer...")

	if err := k.producer.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to shut down the Kafka producer cleanly")
	} else {
		log.Info().Msg("kafka: Closed producer")
	}
}

func createSaramaProducer(cfg *KafkaConfig) (sarama.SyncProducer, error) {
	// Default Sarama config
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.MaxVersion

	// Necessary for SyncProducer
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	// TLS Client auth override
	if cfg.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			return nil, err
		}
		caCert, err := ioutil.ReadFile(cfg.TLS.CaFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}
	}

	// TODO: Find a generic way to override all other configs

	// Build producer
	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return producer, nil
}
