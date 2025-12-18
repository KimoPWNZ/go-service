package analytics

import (
	"context"
	"math"
	"sync"
	"time"

	"go-ai-service/internal/cache"
	"go-ai-service/internal/models"

	"go.uber.org/zap"
)

// AnalyticsService предоставляет сервис аналитики
type AnalyticsService struct {
	mu           sync.RWMutex
	redis        *cache.RedisClient
	stats        *Statistics
	windowSize   int
	threshold    float64
	metricsCache map[string][]models.Metric
	logger       *zap.SugaredLogger
	anomalyChan  chan models.AnalyticsResult
}

// NewAnalyticsService создает новый сервис аналитики
func NewAnalyticsService(redis *cache.RedisClient, windowSize int, threshold float64) *AnalyticsService {
	return &AnalyticsService{
		redis:        redis,
		stats:        NewStatistics(),
		windowSize:   windowSize,
		threshold:    threshold,
		metricsCache: make(map[string][]models.Metric),
		anomalyChan:  make(chan models.AnalyticsResult, 100),
	}
}

// SetLogger устанавливает логгер
func (a *AnalyticsService) SetLogger(logger *zap.SugaredLogger) {
	a.logger = logger
}

// ProcessMetric обрабатывает метрику
func (a *AnalyticsService) ProcessMetric(ctx context.Context, metric models.Metric) (*models.AnalyticsResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	deviceID := metric.DeviceID

	// Добавляем метрику в кэш
	a.metricsCache[deviceID] = append(a.metricsCache[deviceID], metric)

	// Ограничиваем размер окна
	if len(a.metricsCache[deviceID]) > a.windowSize {
		a.metricsCache[deviceID] = a.metricsCache[deviceID][1:]
	}

	// Извлекаем значения RPS для анализа
	rpsValues := make([]float64, len(a.metricsCache[deviceID]))
	for i, m := range a.metricsCache[deviceID] {
		rpsValues[i] = m.RPS
	}

	// Вычисляем статистики
	mean := a.stats.CalculateMean(rpsValues)
	stdDev := a.stats.CalculateStdDev(rpsValues, mean)
	zScore := a.stats.CalculateZScore(metric.RPS, mean, stdDev)

	isAnomaly := math.Abs(zScore) > a.threshold

	result := &models.AnalyticsResult{
		Timestamp:      time.Now(),
		DeviceID:       deviceID,
		RollingAverage: mean,
		StdDev:         stdDev,
		ZScore:         zScore,
		IsAnomaly:      isAnomaly,
		CurrentValue:   metric.RPS,
	}

	// Сохраняем в Redis
	key := "analytics:" + deviceID
	if err := a.redis.Set(ctx, key, result, 5*time.Minute); err != nil {
		a.logger.Errorf("Failed to cache analytics result: %v", err)
	}

	// Отправляем аномалию в канал
	if isAnomaly {
		select {
		case a.anomalyChan <- *result:
			a.logger.Infof("Anomaly detected for device %s: z-score=%.2f", deviceID, zScore)
		default:
			a.logger.Warn("Anomaly channel is full")
		}
	}

	return result, nil
}

// GetAnalytics возвращает аналитику для устройства
func (a *AnalyticsService) GetAnalytics(ctx context.Context, deviceID string) (*models.AnalyticsResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Пытаемся получить из кэша
	key := "analytics:" + deviceID
	var result models.AnalyticsResult
	if err := a.redis.Get(ctx, key, &result); err == nil {
		return &result, nil
	}

	// Если нет в кэше, вычисляем
	metrics, exists := a.metricsCache[deviceID]
	if !exists || len(metrics) == 0 {
		return &models.AnalyticsResult{
			DeviceID:       deviceID,
			RollingAverage: 0,
			StdDev:         0,
			ZScore:         0,
			IsAnomaly:      false,
		}, nil
	}

	rpsValues := make([]float64, len(metrics))
	for i, m := range metrics {
		rpsValues[i] = m.RPS
	}

	latestMetric := metrics[len(metrics)-1]
	mean := a.stats.CalculateMean(rpsValues)
	stdDev := a.stats.CalculateStdDev(rpsValues, mean)
	zScore := a.stats.CalculateZScore(latestMetric.RPS, mean, stdDev)

	result = models.AnalyticsResult{
		Timestamp:      time.Now(),
		DeviceID:       deviceID,
		RollingAverage: mean,
		StdDev:         stdDev,
		ZScore:         zScore,
		IsAnomaly:      math.Abs(zScore) > a.threshold,
		CurrentValue:   latestMetric.RPS,
	}

	return &result, nil
}

// GetAnomalyChannel возвращает канал аномалий
func (a *AnalyticsService) GetAnomalyChannel() <-chan models.AnalyticsResult {
	return a.anomalyChan
}

// GetSummary возвращает сводную статистику
func (a *AnalyticsService) GetSummary() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	summary := make(map[string]interface{})

	var totalMetrics int
	var anomalyCount int

	for deviceID, metrics := range a.metricsCache {
		totalMetrics += len(metrics)

		if len(metrics) > 0 {
			rpsValues := make([]float64, len(metrics))
			for i, m := range metrics {
				rpsValues[i] = m.RPS
			}

			mean := a.stats.CalculateMean(rpsValues)
			stdDev := a.stats.CalculateStdDev(rpsValues, mean)

			if len(rpsValues) > 0 {
				zScore := a.stats.CalculateZScore(rpsValues[len(rpsValues)-1], mean, stdDev)
				if math.Abs(zScore) > a.threshold {
					anomalyCount++
				}
			}
		}
	}

	summary["total_devices"] = len(a.metricsCache)
	summary["total_metrics"] = totalMetrics
	summary["anomaly_count"] = anomalyCount
	summary["window_size"] = a.windowSize
	summary["threshold"] = a.threshold

	return summary
}
