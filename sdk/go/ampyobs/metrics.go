package ampyobs

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg *prometheus.Registry
}

func NewMetrics() *Metrics {
	return &Metrics{reg: prometheus.NewRegistry()}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}

func (m *Metrics) NewCounter(namespace, name, help string, constLabels prometheus.Labels) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, []string{"domain", "outcome", "reason"})
	m.reg.MustRegister(cv)
	return cv
}

func (m *Metrics) NewHistogram(namespace, name, help string, buckets []float64, constLabels prometheus.Labels) *prometheus.HistogramVec {
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		Buckets:     buckets,
		ConstLabels: constLabels,
	}, []string{"domain"})
	m.reg.MustRegister(hv)
	return hv
}

func (m *Metrics) NewGauge(namespace, name, help string, constLabels prometheus.Labels) *prometheus.GaugeVec {
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, []string{"domain"})
	m.reg.MustRegister(gv)
	return gv
}
