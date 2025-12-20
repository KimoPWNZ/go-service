package handlers

import (
	"encoding/json"
	"net/http"

	"go-service/internal/models"

	"go.uber.org/zap"
)

// HealthCheckHandler обработчик для проверки здоровья сервиса
func HealthCheckHandler(logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := models.HealthStatus{
			Status:    "ok",
			Timestamp: "2023-01-01T00:00:00Z", // В реальности использовать time.Now()
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}
}
