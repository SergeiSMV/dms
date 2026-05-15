// Пакет config отвечает за чтение настроек приложения из YAML-файла
// с возможностью переопределения через переменные окружения.
package config

import (
	"fmt"
	"os"
	"strconv"

	// Библиотека для парсинга YAML-файлов в Go-структуры.
	"gopkg.in/yaml.v3"
)

// Config — корневая структура конфигурации приложения.
// Теги `yaml:"..."` указывают, какому ключу в YAML соответствует поле.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	LLM      LLMConfig      `yaml:"llm"`
}

// AuthConfig — настройки аутентификации (JWT-токены).
type AuthConfig struct {
	// Secret — секретный ключ для подписи JWT (HMAC-SHA256). Обязательно менять в продакшене!
	Secret string `yaml:"secret"`
	// AccessTTL — время жизни access-токена в минутах (по умолчанию 15).
	AccessTTL int `yaml:"access_ttl"`
	// RefreshTTL — время жизни refresh-токена в днях (по умолчанию 7).
	RefreshTTL int `yaml:"refresh_ttl"`
}

// ServerConfig — настройки HTTP-сервера.
type ServerConfig struct {
	Port string `yaml:"port"`
}

// LLMConfig — настройки подключения к llm-proxy.
// URL указывает на контейнер llm-proxy, а не напрямую на Ollama.
type LLMConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// DatabaseConfig — настройки подключения к PostgreSQL.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

// DSN собирает строку подключения к PostgreSQL из полей конфигурации.
// Формат: "postgres://user:password@host:port/dbname?sslmode=disable"
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Load читает YAML-файл по указанному пути и возвращает заполненную структуру Config.
// После чтения файла применяются переопределения из переменных окружения.
func Load(path string) (*Config, error) {
	// Читаем содержимое файла целиком в память.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение конфигурации: %w", err)
	}

	cfg := &Config{}

	// yaml.Unmarshal разбирает байты YAML и заполняет поля структуры.
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("парсинг YAML: %w", err)
	}

	// Переменные окружения имеют приоритет над YAML — удобно для Docker/CI.
	applyEnvOverrides(cfg)

	// Значения по умолчанию для auth, если не заданы в YAML.
	if cfg.Auth.Secret == "" {
		cfg.Auth.Secret = "CHANGE-ME-IN-PRODUCTION"
	}
	if cfg.Auth.AccessTTL == 0 {
		cfg.Auth.AccessTTL = 15
	}
	if cfg.Auth.RefreshTTL == 0 {
		cfg.Auth.RefreshTTL = 7
	}

	return cfg, nil
}

// applyEnvOverrides перезаписывает значения конфигурации, если заданы
// соответствующие переменные окружения. Пустая строка = не задана, пропускаем.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.Database.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("LLM_PROXY_URL"); v != "" {
		cfg.LLM.URL = v
	}
	if v := os.Getenv("LLM_PROXY_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("AUTH_SECRET"); v != "" {
		cfg.Auth.Secret = v
	}
	if v := os.Getenv("AUTH_ACCESS_TTL"); v != "" {
		if ttl, err := strconv.Atoi(v); err == nil {
			cfg.Auth.AccessTTL = ttl
		}
	}
	if v := os.Getenv("AUTH_REFRESH_TTL"); v != "" {
		if ttl, err := strconv.Atoi(v); err == nil {
			cfg.Auth.RefreshTTL = ttl
		}
	}
}
