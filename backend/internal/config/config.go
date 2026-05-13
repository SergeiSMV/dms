// Пакет config отвечает за чтение настроек приложения из YAML-файла
// с возможностью переопределения через переменные окружения.
package config

import (
	"fmt"
	"os"

	// Библиотека для парсинга YAML-файлов в Go-структуры.
	"gopkg.in/yaml.v3"
)

// Config — корневая структура конфигурации приложения.
// Теги `yaml:"..."` указывают, какому ключу в YAML соответствует поле.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig — настройки HTTP-сервера.
type ServerConfig struct {
	Port string `yaml:"port"`
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
}
