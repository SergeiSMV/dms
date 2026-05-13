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

	"dms-backend/internal/config"
	"dms-backend/internal/database"
	"dms-backend/internal/server"
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
		// slog.Error + os.Exit(1) — аналог log.Fatal, но через slog.
		slog.Error("не удалось загрузить конфигурацию", "ошибка", err)
		os.Exit(1)
	}
	slog.Info("конфигурация загружена", "порт", cfg.Server.Port)

	// --- База данных ---

	// context.Background() — пустой контекст, корневой для всего приложения.
	dbPool, err := database.Connect(context.Background(), cfg.Database.DSN())
	if err != nil {
		slog.Error("не удалось подключиться к БД", "ошибка", err)
		os.Exit(1)
	}
	// defer — отложенный вызов: pool.Close() выполнится при выходе из main().
	defer dbPool.Close()
	slog.Info("подключение к БД установлено")

	// --- HTTP-сервер ---

	srv := server.New(":" + cfg.Server.Port)

	// Запускаем сервер в отдельной горутине (аналог потока),
	// чтобы main-горутина могла ждать сигнал завершения.
	go func() {
		slog.Info("HTTP-сервер запущен", "адрес", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ошибка HTTP-сервера", "ошибка", err)
			os.Exit(1)
		}
	}()

	// --- Graceful shutdown ---

	// Создаём канал для получения сигналов ОС.
	// SIGINT = Ctrl+C, SIGTERM = docker stop / kill -15.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем main-горутину, пока не придёт сигнал.
	sig := <-quit
	slog.Info("получен сигнал завершения", "сигнал", sig.String())

	// Даём серверу 10 секунд на завершение текущих запросов.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown корректно останавливает сервер: перестаёт принимать новые
	// соединения и ждёт завершения активных запросов.
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("ошибка при остановке сервера", "ошибка", err)
	}

	slog.Info("сервер остановлен")
}
