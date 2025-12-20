package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	initOnce         sync.Once
	goCollector      = prometheus.NewGoCollector()
	processCollector = prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})
)

var (
	RequestsProcessed    *prometheus.CounterVec
	RequestDuration      *prometheus.HistogramVec
	AnomaliesDetected    prometheus.Counter
	MetricsProcessed     prometheus.Counter
	ActiveConnections    prometheus.Gauge
	RedisOperations      *prometheus.CounterVec
	RollingAverageValues prometheus.Histogram
	ZScoreValues         prometheus.Histogram
)

// InitMetrics инициализирует метрики
func InitMetrics() {
	initOnce.Do(func() {
		RequestsProcessed = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "app_requests_processed_total",
				Help: "Total number of processed requests",
			},
			[]string{"endpoint"},
		)

		RequestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "app_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		)

		AnomaliesDetected = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "app_anomalies_detected_total",
				Help: "Total number of anomalies detected",
			},
		)

		MetricsProcessed = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "app_metrics_processed_total",
				Help: "Total number of metrics processed",
			},
		)

		ActiveConnections = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "app_active_connections",
				Help: "Number of active connections",
			},
		)

		RedisOperations = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "app_redis_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"operation"},
		)

		RollingAverageValues = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "app_rolling_average_values",
				Help:    "Values of rolling average calculations",
				Buckets: prometheus.LinearBuckets(0, 500, 20),
			},
		)

		ZScoreValues = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "app_zscore_values",
				Help:    "Values of Z-score calculations",
				Buckets: prometheus.LinearBuckets(-10, 1, 40),
			},
		)

		// Удалены ручные регистрации, чтобы избежать дублирования (promauto регистрирует автоматически)
		// prometheus.MustRegister(goCollector)
		// prometheus.MustRegister(processCollector)
	})
}

// Остальной код без изменений...
func GetHTTPHandler() http.Handler {
	return promhttp.Handler()
}

func RecordRollingAverage(value float64) {
	RollingAverageValues.Observe(value)
}

func RecordZScore(value float64) {
	ZScoreValues.Observe(value)
}

func RecordRedisOperation(operation string) {
	RedisOperations.WithLabelValues(operation).Inc()
}
