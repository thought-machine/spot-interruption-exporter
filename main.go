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
		Help: "The total number of interruption events for a given cluster",
	}, []string{"kubernetes_cluster", "resource_id"})
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	sugar := logger.Sugar()
	cfg, err := LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		sugar.Fatal(err)
	}

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
	sugar.Info("listening for preemption events")
	for event := range e {
		// this ensures we do not handle a duplicate message in the event pubsub sends it more than once
		if exists := c.Exists(event.MessageID); exists {
			continue
		}
		c.Insert(event.MessageID)
		interruptionEvents.WithLabelValues(cfg.ClusterName, event.ResourceID).Inc()
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
