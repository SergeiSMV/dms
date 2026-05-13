package handler

import (
	// Кодирование данных ответа в JSON.
	"encoding/json"
	// Базовые HTTP-типы: Request и ResponseWriter.
	"net/http"
)

// Health — простой endpoint проверки, что сервис запущен и отвечает.
func Health(w http.ResponseWriter, r *http.Request) {
	// Указываем, что ответ будет в формате JSON.
	w.Header().Set("Content-Type", "application/json")
	// Отправляем клиенту JSON: {"status":"ok"}.
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
