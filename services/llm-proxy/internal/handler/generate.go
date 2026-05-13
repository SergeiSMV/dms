package handler

import (
	// Декодирование входного JSON и кодирование JSON-ответов.
	"encoding/json"
	// Базовые HTTP-типы: Request, ResponseWriter, HandlerFunc.
	"net/http"

	// Клиент Ollama и типы запроса/ответа генерации.
	"local_models_api/internal/ollama"
)

// Generate создает хендлер для POST /v1/generate.
// Хендлер валидирует входные данные, вызывает Ollama и возвращает результат.
func Generate(client *ollama.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Читаем JSON-тело запроса в структуру GenerateRequest.
		var req ollama.GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		// Простейшая валидация обязательных полей.
		if req.Model == "" || req.Prompt == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "model and prompt are required"})
			return
		}

		// Передаем запрос в Ollama-клиент (с context текущего HTTP-запроса).
		result, err := client.Generate(r.Context(), req)
		if err != nil {
			// Ошибка внешнего сервиса -> 502 Bad Gateway.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Успешный ответ сгенерированного текста в JSON.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
