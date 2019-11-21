package main

import (
	"flag"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/exporter"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	conf = flag.String("conf", "config.yaml", "The config path file")
)

func main() {
	flag.Parse()

	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.InfoLevel)

	kubeconfig, err := kube.GetKubernetesConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot get kubeconfig")
	}

	b, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read config file")
	}
	var cfg exporter.Config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot parse config to YAML")
	}

	engine := exporter.NewEngine(&cfg, &exporter.ChannelBasedReceiverRegistry{})
	w := kube.NewEventWatcher(kubeconfig, engine.OnEvent)
	w.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	log.Info().Str("signal", sig.String()).Msg("Received signal to exit")
	defer close(c)
	w.Stop()
}
