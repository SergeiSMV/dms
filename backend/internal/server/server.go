// Пакет server создаёт и настраивает HTTP-сервер приложения.
package server

import (
	"net/http"
	"time"

	// chi — легковесный роутер для Go, совместимый со стандартным net/http.
	"github.com/go-chi/chi/v5"
	// Набор готовых middleware: логирование запросов, recovery от panic и др.
	"github.com/go-chi/chi/v5/middleware"
)

// New создаёт и возвращает настроенный HTTP-сервер.
// addr — адрес вида ":8081", на котором сервер будет слушать.
func New(addr string) *http.Server {
	r := chi.NewRouter()

	// --- Middleware (обработчики, которые оборачивают каждый запрос) ---

	// RequestID добавляет уникальный ID к каждому запросу для трассировки.
	r.Use(middleware.RequestID)
	// RealIP извлекает реальный IP клиента из заголовков X-Forwarded-For / X-Real-IP.
	r.Use(middleware.RealIP)
	// Logger выводит в лог метод, путь, статус и время каждого запроса.
	r.Use(middleware.Logger)
	// Recoverer перехватывает panic внутри хендлеров, чтобы сервер не падал.
	r.Use(middleware.Recoverer)
	// Timeout ограничивает время обработки одного запроса (защита от зависаний).
	r.Use(middleware.Timeout(60 * time.Second))

	// --- Маршруты ---

	// Эндпоинт проверки работоспособности сервиса (health check).
	r.Get("/health", handleHealth())

	// http.Server — стандартная структура Go для HTTP-сервера.
	// Настраиваем таймауты чтения/записи для защиты от медленных клиентов.
	return &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
