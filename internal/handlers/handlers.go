package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-service/internal/analytics"
	"go-service/internal/cache"
	"go-service/internal/models"
	"go-service/pkg/metrics"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// RegisterHandlers регистрирует все обработчики
func RegisterHandlers(r *mux.Router, analyticsService *analytics.AnalyticsService, redisClient *cache.RedisClient, logger *zap.SugaredLogger) {
	r.HandleFunc("/health", HealthHandler(logger)).Methods("GET")
	r.HandleFunc("/metrics", MetricsHandler(analyticsService, logger)).Methods("GET")
	r.HandleFunc("/analyze/{deviceID}", AnalyzeHandler(analyticsService, logger)).Methods("GET")
	r.HandleFunc("/metric", MetricHandler(analyticsService, logger)).Methods("POST")

	// Prometheus metrics
	r.Handle("/prometheus", metrics.GetHTTPHandler()).Methods("GET")
}

// HealthHandler обработчик для проверки здоровья
func HealthHandler(logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.RecordRequest("health")
		defer metrics.RecordRequestDuration("health", time.Since(time.Now()))

		response := map[string]string{"status": "ok"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// MetricsHandler обработчик для получения метрик аналитики
func MetricsHandler(analyticsService *analytics.AnalyticsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.RecordRequest("metrics")
		start := time.Now()
		defer metrics.RecordRequestDuration("metrics", time.Since(start))

		summary := analyticsService.GetSummary()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

// AnalyzeHandler обработчик для анализа конкретного устройства
func AnalyzeHandler(analyticsService *analytics.AnalyticsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.RecordRequest("analyze")
		start := time.Now()
		defer metrics.RecordRequestDuration("analyze", time.Since(start))

		vars := mux.Vars(r)
		deviceID := vars["deviceID"]

		result, err := analyticsService.GetAnalytics(r.Context(), deviceID)
		if err != nil {
			logger.Errorf("Failed to get analytics for device %s: %v", deviceID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// MetricHandler обработчик для приема метрик
func MetricHandler(analyticsService *analytics.AnalyticsService, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.RecordRequest("metric")
		start := time.Now()
		defer metrics.RecordRequestDuration("metric", time.Since(start))

		var metric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Errorf("Failed to decode metric: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Валидация
		if metric.DeviceID == "" || metric.RPS < 0 {
			http.Error(w, "Invalid metric data", http.StatusBadRequest)
			return
		}

		result, err := analyticsService.ProcessMetric(r.Context(), metric)
		if err != nil {
			logger.Errorf("Failed to process metric: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		metrics.RecordMetricProcessed()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
