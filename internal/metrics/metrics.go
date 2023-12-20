//go:generate mockery --name Client
package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
)

var (
	interruptionEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "interruption_events_total",
		Help: "The total number of spot interruptions for a given cluster",
	}, []string{"kubernetes_cluster"})
)

// Client provides methods for modifying metrics
type Client interface {
	// IncreaseInterruptionEventCounter increases the interruption metric by one with a label value of cluster
	IncreaseInterruptionEventCounter(cluster string)
	// ServeMetrics serves metrics on the specified port and path of the given
	ServeMetrics(path, port string)
}

func (m *metrics) IncreaseInterruptionEventCounter(cluster string) {
	interruptionEvents.WithLabelValues(cluster).Inc()
}

func (m *metrics) ServeMetrics(path, port string) {
	http.Handle(path, promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()
}

// NewClient creates a new metrics client. To actually start the metrics server, call the Client's ServeMetrics  method.
func NewClient(log *zap.SugaredLogger) Client {
	return &metrics{
		log: log,
	}
}

type metrics struct {
	log *zap.SugaredLogger
}
