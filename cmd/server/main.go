package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ai-service/internal/analytics"
	"go-ai-service/internal/cache"
	"go-ai-service/internal/handlers"
	"go-ai-service/pkg/metrics"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Инициализация Redis
	redisClient := cache.NewRedisClient()
	defer redisClient.Close()

	// Проверка подключения к Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx); err != nil {
		sugar.Errorf("Failed to connect to Redis: %v", err)
		// Вместо os.Exit(1), можно продолжить без Redis или ретраить
		sugar.Warn("Continuing without Redis cache")
	}
	sugar.Info("Connected to Redis successfully")

	// Инициализация аналитики
	analyticsService := analytics.NewAnalyticsService(redisClient, 50, 2.0)

	// Инициализация метрик Prometheus
	metrics.InitMetrics()

	// Создание роутера
	r := mux.NewRouter()

	// Регистрация обработчиков
	handlers.RegisterHandlers(r, analyticsService, redisClient, sugar)

	// Настройка сервера
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		sugar.Infof("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	sugar.Info("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		sugar.Errorf("Server shutdown failed: %v", err)
	}

	sugar.Info("Server stopped")
}
