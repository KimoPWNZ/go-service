package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestsProcessed счетчик обработанных запросов
	RequestsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_requests_processed_total",
			Help: "Total number of processed requests",
		},
		[]string{"endpoint"},
	)

	// RequestDuration гистограмма длительности запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// AnomaliesDetected счетчик обнаруженных аномалий
	AnomaliesDetected = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_anomalies_detected_total",
			Help: "Total number of anomalies detected",
		},
	)

	// MetricsProcessed счетчик обработанных метрик
	MetricsProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "app_metrics_processed_total",
			Help: "Total number of metrics processed",
		},
	)

	// ActiveConnections счетчик активных соединений
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_active_connections",
			Help: "Number of active connections",
		},
	)

	// RedisOperations счетчик операций Redis
	RedisOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation"},
	)

	// RollingAverageValues гистограмма значений скользящего среднего
	RollingAverageValues = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_rolling_average_values",
			Help:    "Values of rolling average calculations",
			Buckets: prometheus.LinearBuckets(0, 500, 20),
		},
	)

	// ZScoreValues гистограмма значений Z-score
	ZScoreValues = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "app_zscore_values",
			Help:    "Values of Z-score calculations",
			Buckets: prometheus.LinearBuckets(-10, 1, 40),
		},
	)
)

// InitMetrics инициализирует метрики
func InitMetrics() {
	// Регистрируем стандартные метрики Go
	prometheus.MustRegister(prometheus.NewGoCollector())
	prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// GetHTTPHandler возвращает обработчик для Prometheus метрик
func GetHTTPHandler() http.Handler {
	return promhttp.Handler()
}

// RecordRollingAverage записывает значение скользящего среднего
func RecordRollingAverage(value float64) {
	RollingAverageValues.Observe(value)
}

// RecordZScore записывает значение Z-score
func RecordZScore(value float64) {
	ZScoreValues.Observe(value)
}

// RecordRedisOperation записывает операцию Redis
func RecordRedisOperation(operation string) {
	RedisOperations.WithLabelValues(operation).Inc()
}
