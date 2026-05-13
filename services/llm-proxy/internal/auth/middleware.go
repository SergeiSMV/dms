package auth

import (
	// Формирование JSON-ответов (например, при ошибке авторизации).
	"encoding/json"
	// Базовые типы HTTP: Request, ResponseWriter, Handler.
	"net/http"
)

// APIKeyMiddleware создает middleware для проверки заголовка X-API-Key.
// validKeys — список разрешенных ключей из конфигурации.
func APIKeyMiddleware(validKeys []string) func(http.Handler) http.Handler {
	// Для быстрых проверок превращаем список ключей в set (map со значением-пустышкой).
	keySet := make(map[string]struct{}, len(validKeys))
	for _, k := range validKeys {
		keySet[k] = struct{}{}
	}

	// Возвращаем стандартный middleware-обработчик в стиле net/http.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Читаем API-ключ из заголовка запроса.
			key := r.Header.Get("X-API-Key")
			// Если ключ не найден среди разрешенных — возвращаем 401 Unauthorized.
			if _, ok := keySet[key]; !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			// Если ключ валиден — передаем запрос следующему обработчику.
			next.ServeHTTP(w, r)
		})
	}
}
