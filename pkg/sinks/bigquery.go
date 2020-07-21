package sinks

import (
	"bufio"
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/batch"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"math/rand"
	"os"
	"time"
	"unicode"
)

// Returns a map filtering out keys that have nil value assigned.
func bigQueryDropNils(x map[string]interface{}) map[string]interface{} {
	y := make(map[string]interface{})
	for key, value := range x {
		if value != nil {
			if mapValue, ok := value.(map[string]interface{}); ok {
				y[key] = bigQueryDropNils(mapValue)
			} else {
				y[key] = value
			}
		}
	}
	return y
}

// Returns a string representing a fixed key. BigQuery expects keys to be valid identifiers, so if they aren't we modify them.
func bigQuerySanitizeKey(key string) string {
	var fixedKey string
	if !unicode.IsLetter(rune(key[0])) {
		fixedKey = "_"
	}
	for _, ch := range key {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			fixedKey = fixedKey + string(ch)
		} else {
			fixedKey = fixedKey + "_"
		}
	}
	return fixedKey
}

// Returns a map copy with fixed keys.
func bigQuerySanitizeKeys(x map[string]interface{}) map[string]interface{} {
	y := make(map[string]interface{})
	for key, value := range x {
		if mapValue, ok := value.(map[string]interface{}); ok {
			y[bigQuerySanitizeKey(key)] = bigQuerySanitizeKeys(mapValue)
		} else {
			y[bigQuerySanitizeKey(key)] = value
		}
	}
	return y
}

func bigQueryWriteBatchToJsonFile(items []interface{}, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for i := 0; i < len(items); i++ {
		event := items[i].(*kube.EnhancedEvent)
		var mapStruct map[string]interface{}
		json.Unmarshal(event.ToJSON(), &mapStruct)
		jsonBytes, _ := json.Marshal(bigQuerySanitizeKeys(bigQueryDropNils(mapStruct)))
		fmt.Fprintln(writer, string(jsonBytes))
	}
	return writer.Flush()
}

func bigQueryCreateDataset(cfg *BigQueryConfig) error {
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, cfg.Project, option.WithCredentialsFile(cfg.CredentialsPath))
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	meta := &bigquery.DatasetMetadata{Location: cfg.Location}
	if err := client.Dataset(cfg.Dataset).Create(ctx, meta); err != nil {
		return err
	}
	return nil
}

func bigQueryImportJsonFromFile(path string, cfg *BigQueryConfig) error {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, cfg.Project, option.WithCredentialsFile(cfg.CredentialsPath))
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	source := bigquery.NewReaderSource(f)
	source.SourceFormat = bigquery.JSON
	source.AutoDetect = true

	loader := client.Dataset(cfg.Dataset).Table(cfg.Table).LoaderFrom(source)
	loader.SchemaUpdateOptions = []string{"ALLOW_FIELD_ADDITION"}

	log.Info().Msgf("BigQuery batch uploading %.3f KBs...", float64(fi.Size())/1e3)
	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	log.Info().Msgf("BigQuery batch uploading done.")
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

type BigQueryConfig struct {
	// BigQuery table config
	Location string `yaml:"location"`
	Project  string `yaml:"project"`
	Dataset  string `yaml:"dataset"`
	Table    string `yaml:"table"`

	// Path to a JSON file that contains your service account key.
	CredentialsPath string `yaml:"credentials_path"`

	// Batching config
	BatchSize       int `yaml:"batch_size"`
	MaxRetries      int `yaml:"max_retries"`
	IntervalSeconds int `yaml:"interval_seconds"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

func NewBigQuerySink(cfg *BigQueryConfig) (*BigQuerySink, error) {
	if cfg.Location == "" {
		cfg.Location = "US"
	}
	if cfg.Project == "" {
		return nil, errors.New("bigquery.project config option must be non-empty")
	}
	if cfg.Dataset == "" {
		return nil, errors.New("bigquery.dataset config option must be non-empty")
	}
	if cfg.Table == "" {
		return nil, errors.New("bigquery.table config option must be non-empty")
	}

	if cfg.BatchSize == 0 {
		cfg.BatchSize = 1000
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.IntervalSeconds == 0 {
		cfg.IntervalSeconds = 10
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 60
	}

	rand.Seed(time.Now().UTC().UnixNano())
	handleBatch := func(ctx context.Context, items []interface{}) []bool {
		res := make([]bool, len(items))
		for i := 0; i < len(items); i++ {
			res[i] = true
		}
		path := fmt.Sprintf("/tmp/bq_batch-%d-%04x.json", time.Now().UTC().Unix(), rand.Uint64()%65535)
		if err := bigQueryWriteBatchToJsonFile(items, path); err != nil {
			log.Error().Msgf("Failed to write JSON file: %v", err)
		}
		if err := bigQueryImportJsonFromFile(path, cfg); err != nil {
			log.Error().Msgf("BigQuerySink load failed: %v", err)
		} else {
			// The batch file is intentionally not deleted in case of failure allowing to manually uplaod it later and debug issues.
			if err := os.Remove(path); err != nil {
				log.Error().Msgf("Failed to delete file %v: %v", path, err)
			}
		}
		return res
	}

	if err := bigQueryCreateDataset(cfg); err != nil {
		log.Error().Msgf("BigQuerySink create dataset failed: %v", err)
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

	return &BigQuerySink{batchWriter: batchWriter}, nil
}

type BigQuerySink struct {
	batchWriter *batch.Writer
}

func (e *BigQuerySink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	e.batchWriter.Submit(ev)
	return nil
}

func (e *BigQuerySink) Close() {
	e.batchWriter.Stop()
}
