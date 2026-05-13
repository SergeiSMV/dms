package main

import (
	// Логирование служебных сообщений и ошибок.
	"log"
	// Базовый HTTP-сервер из стандартной библиотеки Go.
	"net/http"
	// Работа со временем (здесь используется для timeout).
	"time"

	// Chi — легковесный роутер (маршрутизация HTTP-запросов).
	"github.com/go-chi/chi/v5"
	// Набор готовых middleware для логов, recovery, timeout и др.
	"github.com/go-chi/chi/v5/middleware"

	// Загрузка конфигурации из переменных окружения.
	"local_models_api/internal/config"
	// Middleware для авторизации по X-API-Key.
	"local_models_api/internal/auth"
	// HTTP-хендлеры для API-эндпоинтов.
	"local_models_api/internal/handler"
	// HTTP-клиент для взаимодействия с Ollama.
	"local_models_api/internal/ollama"
)

// main — точка входа приложения.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Создаем клиент для запросов к Ollama (URL берется из конфигурации).
	ollamaClient := ollama.NewClient(cfg.OllamaURL)

	// Создаем роутер, который будет обрабатывать входящие HTTP-запросы.
	r := chi.NewRouter()
	// Логирует каждый запрос (метод, путь, статус, время выполнения).
	r.Use(middleware.Logger)
	// Перехватывает panic внутри хендлеров, чтобы сервер не падал целиком.
	r.Use(middleware.Recoverer)
	// Ограничивает максимальное время обработки одного запроса.
	r.Use(middleware.Timeout(6 * time.Minute))

	// Регистрируем GET endpoint для проверки "живости" сервиса.
	r.Get("/health", handler.Health)

	// Группа маршрутов с общим префиксом /v1.
	// Для всех endpoint внутри группы обязателен валидный X-API-Key.
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.APIKeyMiddleware(cfg.APIKeys))
		r.Get("/models", handler.Models(ollamaClient))
		r.Post("/generate", handler.Generate(ollamaClient))
	})

	// Сообщаем в лог, что сервер запускается (порт берётся из ENV).
	log.Printf("server starting on :%s", cfg.Port)
	// Запускаем HTTP-сервер. При фатальной ошибке завершаем программу.
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}

