package models

import (
	"time"
)

// Metric представляет метрику от IoT устройства
type Metric struct {
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`     // Загрузка CPU в процентах
	Memory    float64   `json:"memory"`  // Использование памяти в процентах
	RPS       float64   `json:"rps"`     // Запросов в секунду
	Network   float64   `json:"network"` // Сетевая активность в Мбит/с
}

// AnalyticsResult представляет результат аналитики
type AnalyticsResult struct {
	Timestamp      time.Time `json:"timestamp"`
	DeviceID       string    `json:"device_id"`
	RollingAverage float64   `json:"rolling_average"`
	StdDev         float64   `json:"std_dev"`
	ZScore         float64   `json:"z_score"`
	IsAnomaly      bool      `json:"is_anomaly"`
	CurrentValue   float64   `json:"current_value"`
}

// HealthResponse представляет ответ о состоянии сервиса
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
}

// CacheMetrics представляет метрики кэша
type CacheMetrics struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Size   int64 `json:"size"`
}
