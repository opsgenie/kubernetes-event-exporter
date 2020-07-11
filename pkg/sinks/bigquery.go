package sinks

import (
	"bufio"
	"cloud.google.com/go/bigquery"
	"context"
	"os"
	"fmt"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/batch"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
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

func importJsonFromFile(path, projectID, datasetID, tableID string) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	source := bigquery.NewReaderSource(f)
	source.SourceFormat = bigquery.JSON
	source.AutoDetect = true // Allow BigQuery to determine schema.

	loader := client.Dataset(datasetID).Table(tableID).LoaderFrom(source)

	log.Debug().Msgf("loader.Run...")
	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	log.Debug().Msgf("loader.Wait...")
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	log.Debug().Msgf("loader done.")
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
	// Batching config
	// TODO(vsbus): set default values
	BatchSize       int `yaml:"batch_size"`
	MaxRetries      int `yaml:"max_retries"`
	IntervalSeconds int `yaml:"interval_seconds"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

func NewBigquery(cfg *BigqueryConfig) (*Bigquery, error) {
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
