package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-ai-service/internal/analytics"
	"go-ai-service/internal/cache"
	"go-ai-service/internal/models"
	"go-ai-service/pkg/metrics"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// MetricsHandler обрабатывает запросы метрик
type MetricsHandler struct {
	analyticsService *analytics.AnalyticsService
	redis            *cache.RedisClient
	logger           *zap.SugaredLogger
}

// NewMetricsHandler создает новый MetricsHandler
func NewMetricsHandler(analyticsService *analytics.AnalyticsService, redis *cache.RedisClient, logger *zap.SugaredLogger) *MetricsHandler {
	return &MetricsHandler{
		analyticsService: analyticsService,
		redis:            redis,
		logger:           logger,
	}
}

// Register регистрирует маршруты
func (m *MetricsHandler) Register(router *mux.Router) {
	router.HandleFunc("/metrics", m.HandleMetric).Methods("POST")
	router.HandleFunc("/analytics", m.GetAnalytics).Methods("GET")
	router.HandleFunc("/analytics/{device_id}", m.GetDeviceAnalytics).Methods("GET")
	router.HandleFunc("/summary", m.GetSummary).Methods("GET")
	router.HandleFunc("/cache-metrics", m.GetCacheMetrics).Methods("GET")
}

// HandleMetric обрабатывает входящие метрики
func (m *MetricsHandler) HandleMetric(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Инкрементируем счетчик запросов
	metrics.RequestsProcessed.WithLabelValues("post_metrics").Inc()

	var metric models.Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		metrics.RequestsProcessed.WithLabelValues("error").Inc()
		m.logger.Errorf("Failed to decode metric: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация метрики
	if metric.DeviceID == "" {
		http.Error(w, "DeviceID is required", http.StatusBadRequest)
		return
	}

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// Обработка метрики
	ctx := r.Context()
	result, err := m.analyticsService.ProcessMetric(ctx, metric)
	if err != nil {
		metrics.RequestsProcessed.WithLabelValues("error").Inc()
		m.logger.Errorf("Failed to process metric: %v", err)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	// Инкрементируем счетчик аномалий
	if result.IsAnomaly {
		metrics.AnomaliesDetected.Inc()
	}

	// Сохраняем в Redis для истории
	historyKey := "history:" + metric.DeviceID
	if err := m.redis.Set(ctx, historyKey, metric, 24*time.Hour); err != nil {
		m.logger.Warnf("Failed to save history: %v", err)
	}

	// Формируем ответ
	response := map[string]interface{}{
		"status":       "processed",
		"device_id":    metric.DeviceID,
		"timestamp":    metric.Timestamp,
		"analytics":    result,
		"processed_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// Замеряем время выполнения
	duration := time.Since(start).Seconds()
	metrics.RequestDuration.WithLabelValues("post_metrics").Observe(duration)
}

// GetAnalytics возвращает аналитику
func (m *MetricsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics.RequestsProcessed.WithLabelValues("get_analytics").Inc()

	summary := m.analyticsService.GetSummary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)

	duration := time.Since(start).Seconds()
	metrics.RequestDuration.WithLabelValues("get_analytics").Observe(duration)
}

// GetDeviceAnalytics возвращает аналитику для конкретного устройства
func (m *MetricsHandler) GetDeviceAnalytics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics.RequestsProcessed.WithLabelValues("get_device_analytics").Inc()

	vars := mux.Vars(r)
	deviceID := vars["device_id"]

	if deviceID == "" {
		http.Error(w, "Device ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	result, err := m.analyticsService.GetAnalytics(ctx, deviceID)
	if err != nil {
		m.logger.Errorf("Failed to get analytics: %v", err)
		http.Error(w, "Failed to get analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	duration := time.Since(start).Seconds()
	metrics.RequestDuration.WithLabelValues("get_device_analytics").Observe(duration)
}

// GetSummary возвращает сводную статистику
func (m *MetricsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	metrics.RequestsProcessed.WithLabelValues("get_summary").Inc()

	summary := m.analyticsService.GetSummary()

	// Добавляем метрики Redis
	ctx := r.Context()
	redisMetrics, err := m.redis.GetMetrics(ctx)
	if err == nil {
		summary["redis"] = redisMetrics
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)

	duration := time.Since(start).Seconds()
	metrics.RequestDuration.WithLabelValues("get_summary").Observe(duration)
}

// GetCacheMetrics возвращает метрики кэша
func (m *MetricsHandler) GetCacheMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := m.redis.GetMetrics(ctx)
	if err != nil {
		m.logger.Errorf("Failed to get cache metrics: %v", err)
		http.Error(w, "Failed to get cache metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// RegisterHandlers регистрирует все обработчики
func RegisterHandlers(router *mux.Router, analyticsService *analytics.AnalyticsService, redis *cache.RedisClient, logger *zap.SugaredLogger) {
	// Устанавливаем логгер для сервиса аналитики
	analyticsService.SetLogger(logger)

	// Создаем обработчики
	healthHandler := NewHealthHandler()
	metricsHandler := NewMetricsHandler(analyticsService, redis, logger)

	// Регистрируем маршруты
	healthHandler.Register(router)
	metricsHandler.Register(router)

	// Добавляем маршрут для Prometheus метрик
	router.Handle("/prometheus", metrics.GetHTTPHandler())

	// Middleware для логирования
	router.Use(loggingMiddleware(logger))
}

// loggingMiddleware добавляет логирование запросов
func loggingMiddleware(logger *zap.SugaredLogger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем response writer для захвата статуса
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Infow("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", duration.String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWriter обертка для захвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
