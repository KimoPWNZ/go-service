package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-ai-service/internal/models"

	"github.com/gorilla/mux"
)

// HealthHandler обрабатывает запросы на проверку здоровья
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler создает новый HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// Register регистрирует маршруты
func (h *HealthHandler) Register(router *mux.Router) {
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
	router.HandleFunc("/ready", h.ReadinessCheck).Methods("GET")
	router.HandleFunc("/live", h.LivenessCheck).Methods("GET")
}

// HealthCheck проверяет общее состояние сервиса
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessCheck проверяет готовность сервиса
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// LivenessCheck проверяет живучесть сервиса
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}
