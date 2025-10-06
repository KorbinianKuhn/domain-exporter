package metrics

import (
	"net/http"

	"github.com/korbiniankuhn/domain-exporter/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	checkTimestamp prometheus.Gauge
	checkCounter   prometheus.Counter
	method         *prometheus.GaugeVec
	duration       *prometheus.GaugeVec
	status         *prometheus.GaugeVec
	expiresAt      *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		checkTimestamp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "domain",
				Name:      "check_timestamp_seconds",
				Help:      "Unix timestamp of the last check",
			},
		),
		checkCounter: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "domain",
				Name:      "check_total",
				Help:      "Total number of domain checks",
			},
		),
		method: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "domain",
				Name:      "probe_method",
				Help:      "The method used for probing the domain",
			},
			[]string{"domain"},
		),
		duration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "domain",
				Name:      "probe_duration_seconds",
				Help:      "Duration in seconds of the domain probe",
			},
			[]string{"domain"},
		),
		status: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "domain",
				Name:      "status",
				Help:      "Status of the domain",
			},
			[]string{"domain"},
		),
		expiresAt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "domain",
				Name:      "expires_at",
				Help:      "Expiry date of the domain as a Unix timestamp",
			},
			[]string{"domain"},
		),
	}

	return metrics
}

func (m *Metrics) SetDomainStatus(duration float64, domainInfo *domain.DomainInfo) {
	m.checkCounter.Inc()
	m.checkTimestamp.SetToCurrentTime()

	m.method.WithLabelValues(domainInfo.DomainName).Set(float64(domainInfo.CheckMethod))
	m.duration.WithLabelValues(domainInfo.DomainName).Set(duration)
	m.status.WithLabelValues(domainInfo.DomainName).Set(float64(domainInfo.Status()))
	if domainInfo.ExpiryDate != nil {
		m.expiresAt.WithLabelValues(domainInfo.DomainName).Set(float64(domainInfo.ExpiryDate.Unix()))
	}
}

func (m *Metrics) GetMetricsHandler() http.Handler {
	var r = prometheus.NewRegistry()

	r.MustRegister(
		m.checkTimestamp,
		m.checkCounter,
		m.method,
		m.duration,
		m.status,
		m.expiresAt,
	)

	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	return handler
}
