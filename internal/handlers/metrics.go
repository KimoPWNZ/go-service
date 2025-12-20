package handlers

import (
	"encoding/json"
	"net/http"

	"go-service/internal/analytics"
	"go-service/internal/cache"

	"go.uber.org/zap"
)

// AnalyticsHandler обработчик для аналитики
func AnalyticsHandler(analyticsService *analytics.AnalyticsService, redisClient *cache.RedisClient, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary := analyticsService.GetSummary()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}
