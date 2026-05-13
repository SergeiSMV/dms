package server

import (
	"encoding/json"
	"net/http"
)

// handleHealth возвращает хендлер для GET /health.
// Паттерн «функция возвращает http.HandlerFunc» позволяет в будущем
// передавать зависимости (БД, сервисы) через замыкание (closure).
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем Content-Type до записи тела ответа.
		w.Header().Set("Content-Type", "application/json")
		// json.NewEncoder пишет JSON прямо в ResponseWriter (без промежуточного буфера).
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
