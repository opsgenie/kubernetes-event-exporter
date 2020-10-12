package sinks

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
	"time"
)

type EventBridgeConfig struct {
	DetailType   string              `yaml:"detailTyp"`
	Details   map[string]interface{} `yaml:"details"`
	Source 	  string				`yaml:"source"`
	EventBusName string			`yaml:"eventBusName"`
	Region string				`yaml:"region"`
}

type EventBridgeSink struct {
	cfg *EventBridgeConfig
	svc *eventbridge.EventBridge
}

func NewEventBridgeSink(cfg *EventBridgeConfig) (Sink, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Retryer: client.DefaultRetryer{
				NumMaxRetries:    client.DefaultRetryerMaxNumRetries,
				MinRetryDelay:    client.DefaultRetryerMinRetryDelay,
				MinThrottleDelay: client.DefaultRetryerMinThrottleDelay,
				MaxRetryDelay:     client.DefaultRetryerMaxRetryDelay,
				MaxThrottleDelay: client.DefaultRetryerMaxThrottleDelay,
		},
		},
	)
	if err != nil {
		return nil, err
	}

	svc := eventbridge.New(sess)
	return &EventBridgeSink{
		cfg: cfg,
		svc: svc,
	}, nil
}

func (s *EventBridgeSink) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	log.Info().Msg("Sending event to EventBridge ")
	var toSend string
	if s.cfg.Details != nil {
		res, err := convertLayoutTemplate(s.cfg.Details, ev)
		if err != nil {
			return err
		}

		byte, err := json.Marshal(res)
		toSend = string(byte)
		if err != nil {
			return err
		}
	} else {
		toSend = string(ev.ToJSON())
	}
	tym := time.Now()
	inputRequest := eventbridge.PutEventsRequestEntry{
		Detail: &toSend,
		DetailType: &s.cfg.DetailType,
		Time: &tym,
		Source: &s.cfg.Source,
		EventBusName: &s.cfg.EventBusName,
	}
	log.Info().Str("InputEvent", inputRequest.String()).Msg("Request")
	req,out := s.svc.PutEventsRequest(&eventbridge.PutEventsInput{Entries: []*eventbridge.PutEventsRequestEntry{&inputRequest}})
	err := req.Send()
	if err!=nil{
		log.Error().Str("Failed to send event. Err=", err.Error()).Msg("EventBridge Error")
		log.Error().Str("Failed to send event. Aws out=", out.String()).Msg("EventBridge output")
		return err
	}
	return nil
}

func (s *EventBridgeSink) Close() {
}

