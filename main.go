// Package main listens for interruption events from the specified InterruptionNotifier,
// incrementing a counter every time an event is received
package main

import (
	gcppubsub "cloud.google.com/go/pubsub"
	"context"
	"github.com/thought-machine/spot-interruption-exporter/internal/cache"
	"github.com/thought-machine/spot-interruption-exporter/internal/compute"
	"github.com/thought-machine/spot-interruption-exporter/internal/events"
	"github.com/thought-machine/spot-interruption-exporter/internal/handlers"
	"github.com/thought-machine/spot-interruption-exporter/internal/metrics"
	"go.uber.org/zap"
	"log"
	"os"
	"sync"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("failed to load app configuration: %s", err.Error())
	}

	logger := configureLogger(cfg)
	m := metrics.NewClient(logger)
	m.ServeMetrics(cfg.Prometheus.Path, cfg.Prometheus.Port)

	interruptionEvents, err := createSubscriptionClient(ctx, logger, cfg.Project, cfg.PubSub.InstanceInterruptionSubscriptionName)
	if err != nil {
		logger.Fatal("failed to init instance interruption subscription: %s", err.Error())
	}

	creationEvents, err := createSubscriptionClient(ctx, logger, cfg.Project, cfg.PubSub.InstanceCreationSubscriptionName)
	if err != nil {
		logger.Fatal("failed to init instance creation subscription: %s", err.Error())
	}

	computeClient, err := createComputeClient(ctx, logger, cfg)
	if err != nil {
		logger.Fatal("failed to init compute client")
	}

	initialInstances, err := computeClient.ListInstancesBelongingToKubernetesCluster(ctx)
	if err != nil {
		logger.Fatal("failed to determine initial instances belonging to kubernetes clusters: %s", err.Error())
	}

	interruptions := make(chan *gcppubsub.Message, 30)
	additions := make(chan *gcppubsub.Message, 30)
	instanceToClusterMappings := cache.NewCacheWithTTLFrom(cache.NoExpiration, initialInstances)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go interruptionEvents.Receive(ctx, interruptions)
	go creationEvents.Receive(ctx, additions)
	logger.Info("listening for instance creation & interruption events")

	go handlers.HandleInterruptionEvents(interruptions, instanceToClusterMappings, m, logger, wg)
	go handlers.HandleCreationEvents(additions, instanceToClusterMappings, logger, wg)
	logger.Info("handlers started for instance creation & interruption events")
	wg.Wait()
}

func createComputeClient(ctx context.Context, log *zap.SugaredLogger, cfg Config) (compute.Client, error) {
	return compute.NewClient(ctx, compute.NewClientInput{
		Logger:    log,
		ProjectID: cfg.Project,
	})
}

func createSubscriptionClient(ctx context.Context, log *zap.SugaredLogger, projectID, subscriptionName string) (events.Subscription, error) {
	return events.NewPubSubNotifier(ctx, &events.PubSubNotifierInput{
		Logger:           log,
		ProjectID:        projectID,
		SubscriptionName: subscriptionName,
	})
}

func configureLogger(cfg Config) *zap.SugaredLogger {
	loggerConfig := zap.NewProductionConfig()
	if err := configureLogLevel(&loggerConfig, cfg.LogLevel); err != nil {
		log.Fatalf("failed to parse log level: %s", err.Error())
	}
	loggerConfig.EncoderConfig.TimeKey = "time"
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	return logger.Sugar()
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
