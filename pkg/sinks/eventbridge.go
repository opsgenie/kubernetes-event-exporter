package sinks

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"log"
	"time"
)

const retries = 3

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
		Region: aws.String(cfg.Region)},
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
	log.Println("[INFO] Sending event to EventBridge ")
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
	inputRequest := eventbridge.PutEventsRequestEntry{
		Detail: &toSend,
		DetailType:,
		Time: ,
		Source: ,
		EventBusName: ,
	}
	req,out := s.svc.PutEventsRequest(&eventbridge.PutEventsInput{Entries: []*eventbridge.PutEventsRequestEntry{&inputRequest}})
	err := req.Send()
	var retry int64
	retry = 0
	for err!=nil && retry<=retries {
		log.Printf("[WARN] err = %v \n, Retrying SendEvents to EventBridge \n", err)
		time.Sleep(time.Second * time.Duration(retry+1))
		retry = retry + 1
		err = req.Send()
	}
	if err!=nil{
		log.Printf("[ERROR] Failed to send event. Err = %v \n", err)
		log.Printf("[ERROR] Event = %v \n", out)
		return err
	}
	return nil
}

func (s *EventBridgeSink) Close() {
}

