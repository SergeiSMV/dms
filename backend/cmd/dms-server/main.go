// Точка входа DMS-сервера.
// Читает конфигурацию, подключается к БД, запускает HTTP-сервер
// и корректно завершается при получении сигнала ОС (Ctrl+C, docker stop).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dms-backend/internal/auth"
	"dms-backend/internal/config"
	"dms-backend/internal/database"
	"dms-backend/internal/repository"
	"dms-backend/internal/server"
	"dms-backend/internal/service"
)

func main() {
	// slog — структурированный логгер из стандартной библиотеки Go 1.21+.
	// TextHandler выводит логи в формате key=value (удобно для разработки).
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	// Устанавливаем как логгер по умолчанию для всего приложения.
	slog.SetDefault(logger)

	// --- Конфигурация ---

	// Путь к YAML-файлу можно переопределить через переменную окружения.
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("не удалось загрузить конфигурацию", "ошибка", err)
		os.Exit(1)
	}
	slog.Info("конфигурация загружена", "порт", cfg.Server.Port)

	// --- База данных ---

	ctx := context.Background()

	dbPool, err := database.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		slog.Error("не удалось подключиться к БД", "ошибка", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("подключение к БД установлено")

	// --- Миграции ---

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	if err := database.RunMigrations(cfg.Database.DSN(), migrationsPath); err != nil {
		slog.Error("не удалось применить миграции", "ошибка", err)
		os.Exit(1)
	}

	// --- Seed: организация и admin при первом запуске ---

	seedResult, err := database.Seed(ctx, dbPool)
	if err != nil {
		slog.Error("не удалось выполнить seed", "ошибка", err)
		os.Exit(1)
	}

	// --- Инициализация слоёв ---

	// JWT-менеджер с настройками из конфигурации.
	jwtManager := auth.NewJWTManager(
		cfg.Auth.Secret,
		time.Duration(cfg.Auth.AccessTTL)*time.Minute,
		time.Duration(cfg.Auth.RefreshTTL)*24*time.Hour,
	)

	// Репозиторий → Сервис → Сервер (слои зависят только «вниз»).
	userRepo := repository.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, jwtManager)

	// --- HTTP-сервер ---

	srv := server.New(":"+cfg.Server.Port, server.Deps{
		AuthService:  authService,
		JWTManager:   jwtManager,
		DefaultOrgID: seedResult.OrgID,
	})

	// Запускаем сервер в отдельной горутине.
	go func() {
		slog.Info("HTTP-сервер запущен", "адрес", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ошибка HTTP-сервера", "ошибка", err)
			os.Exit(1)
		}
	}()

	// --- Graceful shutdown ---

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	slog.Info("получен сигнал завершения", "сигнал", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("ошибка при остановке сервера", "ошибка", err)
	}

	slog.Info("сервер остановлен")
}
