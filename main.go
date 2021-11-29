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

	log.Logger = log.With().Caller().Logger().Level(zerolog.DebugLevel)

	if cfg.LogLevel != "" {
		level, err := zerolog.ParseLevel(cfg.LogLevel)
		if err != nil {
			log.Fatal().Err(err).Str("level", cfg.LogLevel).Msg("Invalid log level")
		}
		log.Logger = log.Logger.Level(level)
	}

	if cfg.LogFormat == "json" {
		// Defaults to JSON already nothing to do
	} else if cfg.LogFormat == "" || cfg.LogFormat == "pretty" {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		})
	} else {
		log.Fatal().Str("log_format", cfg.LogFormat).Msg("Unknown log format")
	}

	if cfg.ThrottlePeriod == 0 {
		cfg.ThrottlePeriod = 5
	}

	kubeconfig, err := kube.GetKubernetesConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot get kubeconfig")
	}

	engine := exporter.NewEngine(&cfg, &exporter.ChannelBasedReceiverRegistry{})
	onEvent := engine.OnEvent
	if len(cfg.ClusterName) != 0 {
		onEvent = func(event *kube.EnhancedEvent) {
			// note that per code this value is not set anywhere on the kubernetes side
			// https://github.com/kubernetes/apimachinery/blob/v0.22.4/pkg/apis/meta/v1/types.go#L276
			event.ClusterName = cfg.ClusterName
			engine.OnEvent(event)
		}
	}
	w := kube.NewEventWatcher(kubeconfig, cfg.Namespace, cfg.ThrottlePeriod, onEvent)

	ctx, cancel := context.WithCancel(context.Background())
	leaderLost := make(chan bool)
	if cfg.LeaderElection.Enabled {
		l, err := kube.NewLeaderElector(cfg.LeaderElection.LeaderElectionID, kubeconfig,
			func(_ context.Context) {
				log.Info().Msg("leader election got")
				w.Start()
			},
			func() {
				log.Error().Msg("leader election lost")
				leaderLost <- true
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

	gracefulExit := func() {
		defer close(c)
		defer close(leaderLost)
		cancel()
		w.Stop()
		engine.Stop()
		log.Info().Msg("Exiting")
	}

	select {
	case sig := <-c:
		log.Info().Str("signal", sig.String()).Msg("Received signal to exit")
		gracefulExit()
	case <-leaderLost:
		log.Warn().Msg("Leader election lost")
		gracefulExit()
	}
}
