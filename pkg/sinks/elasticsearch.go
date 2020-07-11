package sinks

import (
	// "bytes"
	"bufio"
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"os"
	// "crypto/tls"
	// "encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	// "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/batch"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
	// "io/ioutil"
	// "net/http"
	"regexp"
	"strings"
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
		fmt.Fprintln(writer, string(items[i].ToJSON()))
	}
	return writer.Flush()
}

func importJsonFromFile(path string) error {
	projectID := "autonomous-173023"
	datasetID := "av_viktor"
	tableID := "k8s_test_08"
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

	log.Info().Msgf("loader.Run...")
	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	log.Info().Msgf("loader.Wait...")
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	log.Info().Msgf("loader done.")
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}

type ElasticsearchConfig struct {
	// Connection specific
	Hosts    []string `yaml:"hosts"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	CloudID  string   `yaml:"cloudID"`
	APIKey   string   `yaml:"apiKey"`
	// Indexing preferences
	UseEventID  bool   `yaml:"useEventID"`
	Index       string `yaml:"index"`
	IndexFormat string `yaml:"indexFormat"`
	TLS         struct {
		InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
		ServerName         string `yaml:"serverName"`
		CaFile             string `yaml:"caFile"`
	} `yaml:"tls"`
	Layout map[string]interface{} `yaml:"layout"`
}

func NewElasticsearch(cfg *ElasticsearchConfig) (*Elasticsearch, error) {
	fmt.Println("NewElasticsearch")
	log.Info().Msgf("NewElasticsearch cfg: %v", cfg)
	fmt.Println("NewElasticsearch")
	// var caCert []byte

	// if len(cfg.TLS.CaFile) > 0 {
	// 	readFile, err := ioutil.ReadFile(cfg.TLS.CaFile)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	caCert = readFile
	// }

	// tlsClientConfig := &tls.Config{
	// 	InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	// 	ServerName:         cfg.TLS.ServerName,
	// }
	// tlsClientConfig.RootCAs.AppendCertsFromPEM(caCert)

	// client, err := elasticsearch.NewClient(elasticsearch.Config{
	// 	Addresses: cfg.Hosts,
	// 	Username:  cfg.Username,
	// 	Password:  cfg.Password,
	// 	CloudID:   cfg.CloudID,
	// 	APIKey:    cfg.APIKey,
	// 	Transport: &http.Transport{
	// 		TLSClientConfig: tlsClientConfig,
	// 	},
	// })
	// if err != nil {
	// 	return nil, err
	// }

	myfunc := func(ctx context.Context, items []interface{}) []bool {
		res := make([]bool, len(items))
		for i := 0; i < len(items); i++ {
			res[i] = true
		}
		path := "/tmp/batch.json"
		if err := writeBatchToJsonFile(items, path); err != nil {
			log.Error().Msgf("Failed to write JSON file: %v", err)
		}
		if err := importJsonFromFile(path); err != nil {
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
			BatchSize:  1000,
			MaxRetries: 3,
			Interval:   time.Duration(10) * time.Second,
			Timeout:    time.Duration(60) * time.Second,
		},
		myfunc,
	)
	batchWriter.Start()
	return &Elasticsearch{
		client:      nil,
		cfg:         nil,
		batchWriter: batchWriter,
	}, nil
}

type Elasticsearch struct {
	client      *elasticsearch.Client
	cfg         *ElasticsearchConfig
	batchWriter *batch.Writer
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
	log.Info().Msgf("add to buffer...")
	e.batchWriter.Submit(ev)
	return nil
	// var index string
	// if len(e.cfg.IndexFormat) > 0 {
	// 	now := time.Now()
	// 	index = formatIndexName(e.cfg.IndexFormat, now)
	// } else {
	// 	index = e.cfg.Index
	// }

	// req := esapi.IndexRequest{
	// 	Body:  bytes.NewBuffer(b),
	// 	Index: index,
	// }

	// if e.cfg.UseEventID {
	// 	req.DocumentID = string(ev.UID)
	// }

	// resp, err := req.Do(ctx, e.client)
	// if err != nil {
	// 	return err
	// }

	// defer resp.Body.Close()
	// _ = resp.Body
	// return nil
}

func (e *Elasticsearch) Close() {
	// No-op
}
