package sinks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
	"net"
)

type UDPConfig struct {
	Host    string `yaml:"host"`
	Layout map[string]interface{} `yaml:"layout"`
}

type UDPClient struct {
	msgChan chan []byte
	context context.Context
	conn net.Conn
	cfg *UDPConfig
}

func NewUDPClient(cfg *UDPConfig)(*UDPClient, error){
	raddr, err := net.ResolveUDPAddr("udp", cfg.Host)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	client := &UDPClient{
		msgChan: make(chan []byte),
		context: context.Background(),
		conn: conn,
		cfg: cfg,
	}

	go client.loop()

	return client, nil
}

func (u *UDPClient) loop(){
	for {
		select {
		case msg := <-u.msgChan:
			_, err := fmt.Fprintf(u.conn, "%s", msg)
			if err != nil {
				log.Debug().Str("sink", "udp").Str("event", err.Error()).Msg("failed to send log")
			}
			log.Debug().Str("sink", "udp").Msg("log send successfully")
		case <-u.context.Done():
			log.Debug().Str("sink", "udp").Msg("context done")
			return
		}
	}
}

func (u *UDPClient) Send(ctx context.Context, ev *kube.EnhancedEvent) error {
	var toSend []byte

	if u.cfg.Layout != nil {
		res, err := convertLayoutTemplate(u.cfg.Layout, ev)
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

	u.msgChan <- append(toSend, '\n')
	return nil
}

func (u *UDPClient) Close() {
	u.conn.Close()
}
