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

type Client interface {
	IncreaseInterruptionEventCounter(cluster string)
	RegisterHTTPHandler()
}

func (m *metrics) IncreaseInterruptionEventCounter(cluster string) {
	interruptionEvents.WithLabelValues(cluster).Inc()
}

func (m *metrics) RegisterHTTPHandler() {
	http.Handle(m.path, promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", m.port), nil))
	}()
}

func NewClient(path, port string, log *zap.SugaredLogger) Client {
	return &metrics{
		path: path,
		port: port,
		log:  log,
	}
}

type metrics struct {
	path string
	port string
	log  *zap.SugaredLogger
}
