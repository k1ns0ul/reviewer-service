package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reviewer-service/internal/application/service"
	httpHandler "reviewer-service/internal/entrypoints/http"
	"reviewer-service/internal/infrastructure/repository"
	"reviewer-service/pkg/database"
	"reviewer-service/pkg/logger"
)

func main() {
	appLogger := logger.NewDefault()
	appLogger.Info("запуск reviewer service")

	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "reviewer_service"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	serverPort := getEnv("SERVER_PORT", "8080")

	appLogger.Info("подключение к базе данных",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"database", dbConfig.DBName,
	)

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		appLogger.Error("ошибка подключения к базе данных", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	appLogger.Info("соединение с базой данных установлено")

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	prRepo := repository.NewPRRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	router := httpHandler.NewRouter(appLogger, userService, teamService, prService, statsRepo)

	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("запуск HTTP сервера", "port", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("ошибка сервера", "error", err)
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("принудительная остановка сервера", "error", err)
		log.Fatal(err)
	}

	appLogger.Info("сервер остановлен")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
		return value
	}
	return defaultValue
}
