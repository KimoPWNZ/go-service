package models

// HealthStatus статус здоровья (если отличается от HealthResponse в metric.go)
type HealthStatus struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}
