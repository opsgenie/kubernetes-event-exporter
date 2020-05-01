package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opsgenie/kubernetes-event-exporter/pkg/exporter"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var (
	conf = flag.String("conf", "config.yaml", "The config path file")
)

func main() {
	flag.Parse()
	b, err := ioutil.ReadFile(*conf)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot read config file")
	}

	b = []byte(os.ExpandEnv(string(b)))

	var cfg exporter.Config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot parse config to YAML")
	}

	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.DebugLevel)

	if cfg.LogLevel != "" {
		level, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Fatal().Err(err).Str("level", cfg.LogLevel).Msg("Invalid log level")
		}
		log.Logger = log.Logger.Level(level)
	}

	kubeconfig, err := kube.GetKubernetesConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot get kubeconfig")
	}

	engine := exporter.NewEngine(&cfg, &exporter.ChannelBasedReceiverRegistry{})
	w := kube.NewEventWatcher(kubeconfig, engine.OnEvent)

	ctx, cancel := context.WithCancel(context.Background())
	if cfg.LeaderElection.Enabled {
		l, err := kube.NewLeaderElector(cfg.LeaderElection.LeaderElectionID, kubeconfig,
			func(_ context.Context) {
				log.Info().Msg("leader election got")
				w.Start()
			},
			func() {
				log.Error().Msg("leader election lost")
				w.Stop()
			},
		)
		if err != nil {
			log.Fatal().Err(err).Msg("create leaderelector failed")
		}
		go l.Run(ctx)
	} else {
		w.Start()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	log.Info().Str("signal", sig.String()).Msg("Received signal to exit")
	defer close(c)
	if cfg.LeaderElection.Enabled {
		cancel()
	} else {
		w.Stop()
	}
	engine.Stop()
	log.Info().Msg("Exiting")
}
