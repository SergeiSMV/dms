package config

import (
	// Формирование и возврат ошибок.
	"errors"
	// Чтение переменных окружения.
	"os"
	// Работа со строками (split/trim).
	"strings"

	// Загрузка переменных из .env файла в окружение процесса.
	"github.com/joho/godotenv"
)

// Config хранит все runtime-настройки приложения.
type Config struct {
	// Порт, на котором запускается HTTP-сервер (например, "8080").
	Port      string
	// URL инстанса Ollama, к которому будет обращаться сервис.
	OllamaURL string
	// Список валидных API-ключей для доступа к защищенным endpoint.
	APIKeys   []string
}

// Load читает конфигурацию из переменных окружения и возвращает готовую структуру Config.
func Load() (*Config, error) {
	// Пытаемся загрузить .env (если файла нет, это не критично).
	godotenv.Load()

	// PORT необязателен: если не задан, используем значение по умолчанию.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// OLLAMA_URL обязателен: без него неизвестно, куда отправлять запросы к модели.
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		return nil, errors.New("OLLAMA_URL is required")
	}

	// API_KEYS обязателен: без ключей нельзя проверять доступ к защищенным маршрутам.
	apiKeysRaw := os.Getenv("API_KEYS")
	if apiKeysRaw == "" {
		return nil, errors.New("API_KEYS is required")
	}

	// Разбиваем строку вида "key1,key2,key3" в срез строк.
	keys := strings.Split(apiKeysRaw, ",")
	// Удаляем лишние пробелы вокруг каждого ключа.
	for i, k := range keys {
		keys[i] = strings.TrimSpace(k)
	}

	// Собираем и возвращаем итоговую конфигурацию приложения.
	return &Config{
		Port:      port,
		OllamaURL: ollamaURL,
		APIKeys:   keys,
	}, nil
}
