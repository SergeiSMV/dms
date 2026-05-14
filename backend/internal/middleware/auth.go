// Пакет middleware содержит HTTP-middleware для chi-роутера.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"dms-backend/internal/auth"
)

// contextKey — тип для ключей контекста, чтобы избежать коллизий со строками.
type contextKey string

const (
	// Ключи для хранения данных аутентификации в context запроса.
	UserIDKey contextKey = "user_id"
	OrgIDKey  contextKey = "org_id"
	RoleKey   contextKey = "role"
)

// Auth возвращает middleware, проверяющий JWT в заголовке Authorization.
// Формат: "Bearer <token>". При невалидном токене — 401 Unauthorized.
func Auth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем заголовок Authorization.
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"отсутствует токен авторизации"}`, http.StatusUnauthorized)
				return
			}

			// Отрезаем префикс "Bearer ".
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"неверный формат токена"}`, http.StatusUnauthorized)
				return
			}

			// Парсим и проверяем JWT.
			claims, err := jwtManager.ParseToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"невалидный токен"}`, http.StatusUnauthorized)
				return
			}

			// Принимаем только access-токены (не refresh).
			if claims.Type != auth.AccessToken {
				http.Error(w, `{"error":"требуется access-токен"}`, http.StatusUnauthorized)
				return
			}

			// Кладём данные пользователя в контекст запроса.
			// Все последующие хендлеры смогут их достать через GetUserID(ctx) и т.д.
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, OrgIDKey, claims.OrgID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole возвращает middleware, проверяющий роль пользователя.
// Если роль не входит в allowed — 403 Forbidden.
func RequireRole(allowed ...string) func(http.Handler) http.Handler {
	// Используем map для быстрой проверки вхождения.
	roleSet := make(map[string]bool, len(allowed))
	for _, role := range allowed {
		roleSet[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetRole(r.Context())
			if !roleSet[role] {
				http.Error(w, `{"error":"недостаточно прав"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID извлекает ID пользователя из контекста запроса.
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

// GetOrgID извлекает ID организации из контекста запроса.
func GetOrgID(ctx context.Context) string {
	v, _ := ctx.Value(OrgIDKey).(string)
	return v
}

// GetRole извлекает роль пользователя из контекста запроса.
func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(RoleKey).(string)
	return v
}
