package handler

import (
	// Кодирование структур/срезов в JSON для HTTP-ответа.
	"encoding/json"
	// Базовые HTTP-типы: Request, ResponseWriter, HandlerFunc.
	"net/http"

	// Клиент для обращения к Ollama API.
	"local_models_api/internal/ollama"
)

// Models создает HTTP-хендлер, который возвращает список моделей из Ollama.
func Models(client *ollama.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Запрашиваем список моделей у внешнего сервиса Ollama.
		models, err := client.ListModels(r.Context())
		if err != nil {
			// Если Ollama недоступна/вернула ошибку — отдаём 502 Bad Gateway.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		// При успехе возвращаем список моделей в формате JSON.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}
}
