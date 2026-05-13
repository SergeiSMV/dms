package ollama

import (
	// Формирование тела POST-запроса из байт.
	"bytes"
	// Context нужен для отмены запроса и соблюдения deadline от вызывающего кода.
	"context"
	// Декодирование JSON-ответов от Ollama.
	"encoding/json"
	// Форматирование ошибок с контекстом.
	"fmt"
	// HTTP-клиент и связанные типы запросов/ответов.
	"net/http"
	// Настройка timeout для исходящих запросов.
	"time"
)

// Client — минимальный HTTP-клиент для работы с Ollama API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает клиент Ollama с базовым URL и общим timeout 5 минут.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Model — публичное представление модели, которое отдаем наружу.
type Model struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Внутренняя структура для разбора ответа Ollama /api/tags.
type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

// ListModels получает список доступных моделей из Ollama.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	// Формируем GET-запрос к endpoint /api/tags с переданным context.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// Отправляем запрос в Ollama.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	// Ожидаем успешный статус 200 OK.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Декодируем JSON-ответ в промежуточную структуру.
	var tags tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Преобразуем внутренний формат ответа в публичный тип []Model.
	models := make([]Model, len(tags.Models))
	for i, m := range tags.Models {
		models[i] = Model{Name: m.Name, Size: m.Size}
	}
	return models, nil
}

// GenerateRequest — входные данные для генерации текста.
type GenerateRequest struct {
	// Имя модели в Ollama (например, "qwen2.5:14b-instruct-q4_K_M").
	Model  string `json:"model"`
	// Текст запроса (промпт), который отправляется модели.
	Prompt string `json:"prompt"`
}

// GenerateResponse — сокращенный ответ от /api/generate, который возвращаем дальше по API.
type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Внутренний формат запроса к Ollama /api/generate.
// Отдельная структура нужна, чтобы явно контролировать поля запроса (например, stream=false).
type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Generate отправляет промпт в Ollama и возвращает сгенерированный ответ модели.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// Сериализуем тело JSON для POST /api/generate.
	body, err := json.Marshal(ollamaGenerateRequest{Model: req.Model, Prompt: req.Prompt, Stream: false})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Формируем HTTP-запрос с context, чтобы поддерживать отмену/таймауты.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Выполняем запрос в Ollama.
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	// Если Ollama вернула не 200, пробуем прочитать текст ошибки из тела ответа.
	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != "" {
			return nil, fmt.Errorf("ollama: %s", errBody.Error)
		}
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Декодируем успешный JSON-ответ и возвращаем его вызывающему коду.
	var ollamaResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &ollamaResp, nil
}
