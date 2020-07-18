package sinks

import (
	"bufio"
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/batch"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"os"
	"time"
)

func writeBatchToJsonFile(items []interface{}, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for i := 0; i < len(items); i++ {
		event := items[i].(*kube.EnhancedEvent)
		fmt.Fprintln(writer, string(event.ToJSON()))
	}
	return writer.Flush()
}

func (e *Bigquery) createDataset() error {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, e.Project, option.WithCredentialsFile(e.CredentialsPath))
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	meta := &bigquery.DatasetMetadata{Location: "US"}
	if err := client.Dataset(e.Dataset).Create(ctx, meta); err != nil {
		return err
	}
	return nil
}

func (e *Bigquery) importJsonFromFile(path string) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, e.Project, option.WithCredentialsFile(e.CredentialsPath))
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	source := bigquery.NewReaderSource(f)
	source.SourceFormat = bigquery.JSON
	source.AutoDetect = true // Allow BigQuery to determine schema.

	loader := client.Dataset(e.Dataset).Table(e.Table).LoaderFrom(source)

	log.Info().Msgf("Bigquery batch uploading %f MBs...", float64(fi.Size())/1e6)
	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	log.Info().Msgf("Bigquery batch uploading done.")
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

// TODO(vsbus): test it with a table that has limited permissions
type BigqueryConfig struct {
	// BigQuery table config
	// TODO(vsbus): add validator for BQ configs to be set
	Project string `yaml:"project"`
	Dataset string `yaml:"dataset"`
	Table   string `yaml:"table"`

	// Path to a JSON file that contains your service account key.
	CredentialsPath string `yaml:"credentials_path"`

	// Batching config
	// TODO(vsbus): set default values
	BatchSize       int `yaml:"batch_size"`
	MaxRetries      int `yaml:"max_retries"`
	IntervalSeconds int `yaml:"interval_seconds"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

func NewBigquery(cfg *BigqueryConfig) (*Bigquery, error) {
	if cfg.Project == "" {
		return nil, "BigqueryConfig.Project must be non-empty"
	}
	if cfg.Dataset == "" {
		return nil, "BigqueryConfig.Dataset must be non-empty"
	}
	if cfg.Table == "" {
		return nil, "BigqueryConfig.Dataset must be non-empty"
	}

	if cfg.BatchSize == 0 {
		return nil, "BigqueryConfig.BatchSize must be positive"
	}
	if cfg.MaxRetries == 0 {
		return nil, "BigqueryConfig.MaxRetries must be positive"
	}
	if cfg.IntervalSeconds == 0 {
		return nil, "BigqueryConfig.IntervalSeconds must be positive"
	}
	if cfg.TimeoutSeconds == 0 {
		return nil, "BigqueryConfig.TimeoutSeconds must be positive"
	}

	handleBatch := func(ctx context.Context, items []interface{}) []bool {
		res := make([]bool, len(items))
		for i := 0; i < len(items); i++ {
			res[i] = true
		}
		path := "/tmp/bigquery_batch.json"
		if err := writeBatchToJsonFile(items, path); err != nil {
			log.Error().Msgf("Failed to write JSON file: %v", err)
		}
		if err := importJsonFromFile(path, cfg.Project, cfg.Dataset, cfg.Table); err != nil {
			log.Error().Msgf("BigQuery load failed: %v", err)
		} else {
			if err := os.Remove(path); err != nil {
				log.Error().Msgf("Failed to delete file %v: %v", path, err)
			}
		}
		return res
	}

	if err := createDataset(cfg.Project, cfg.Dataset); err != nil {
		log.Error().Msgf("BigQuery create dataset failed: %v", err)
	}

	batchWriter := batch.NewWriter(
		batch.WriterConfig{
			BatchSize:  cfg.BatchSize,
			MaxRetries: cfg.MaxRetries,
			Interval:   time.Duration(cfg.IntervalSeconds) * time.Second,
			Timeout:    time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
		handleBatch,
	)
	batchWriter.Start()

	return &Bigquery{batchWriter: batchWriter}, nil
}

type Bigquery struct {
	batchWriter *batch.Writer
}

func (e *Bigquery) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	e.batchWriter.Submit(ev)
	return nil
}

func (e *Bigquery) Close() {
	// No-op
}
