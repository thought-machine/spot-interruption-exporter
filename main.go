// Package main listens for interruption events from the specified Notifier,
// incrementing a counter every time an event is received
package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/thought-machine/spot-interruption-exporter/internal/cache"
	"github.com/thought-machine/spot-interruption-exporter/internal/events"
)

var (
	interruptionEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "interruption_events_total",
		Help: "The total number of spot interruptions for a given cluster",
	}, []string{"kubernetes_cluster"})
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	loggerConfig := zap.NewProductionConfig()
	if err := configureLogLevel(&loggerConfig, cfg.LogLevel); err != nil {
		panic(fmt.Sprintf("failed to parse log level: %s", err.Error()))
	}
	loggerConfig.EncoderConfig.TimeKey = "time"
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	sugar := logger.Sugar()
	http.Handle(cfg.Prometheus.Path, promhttp.Handler())
	go func() {
		sugar.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Prometheus.Port), nil))
	}()

	n, err := createCSPNotifier(ctx, sugar, cfg)
	if err != nil {
		sugar.Fatal("failed to init cloud-provider client: %s", err.Error())
	}

	e := make(chan events.InterruptionEvent)
	go n.Receive(ctx, e)

	c := cache.NewCacheWithTTL(time.Minute * 10)
	for event := range e {
		s := sugar.With("message_id", event.MessageID, "resource_id", event.ResourceID)
		// this ensures we do not handle a duplicate message in the event pubsub sends it more than once
		if exists := c.Exists(event.MessageID); exists {
			s.Debug("handled duplicate message")
			continue
		}
		c.Insert(event.MessageID)
		interruptionEvents.WithLabelValues(cfg.ClusterName).Inc()
		s.Info("interrupted")
	}
}

func createCSPNotifier(ctx context.Context, log *zap.SugaredLogger, cfg Config) (events.Notifier, error) {
	switch {
	case strings.EqualFold(cfg.Provider, "gcp"):
		n, err := events.NewPubSubNotifier(ctx, &events.NewPubSubNotifierInput{
			Logger:           log,
			ProjectID:        cfg.GCP.Project,
			SubscriptionName: cfg.GCP.PubSubSubscriptionName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP client: %w", err)
		}
		return n, nil
	default:
		return nil, fmt.Errorf("unknown or unsupported cloud provider: %s", cfg.Provider)
	}
}

func configureLogLevel(lCfg *zap.Config, logLevel string) error {
	if len(logLevel) == 0 {
		return nil
	}

	l, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return err
	}
	lCfg.Level = l
	return nil
}
