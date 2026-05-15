// Пакет middleware содержит HTTP-middleware для chi-роутера.
package middleware

import (
	"net/http"
	"strings"

	"dms-backend/internal/appctx"
	"dms-backend/internal/auth"
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

			// Проверяем, что в токене есть обязательные поля.
			// Без org_id запросы к БД вернут пустые результаты или сломают фильтрацию.
			if claims.UserID == "" || claims.OrgID == "" {
				http.Error(w, `{"error":"невалидный токен: отсутствует user_id или org_id"}`, http.StatusUnauthorized)
				return
			}

			// Кладём данные пользователя в контекст через appctx.
			// Репозитории и сервисы читают эти значения для фильтрации по организации.
			ctx := appctx.WithUserID(r.Context(), claims.UserID)
			ctx = appctx.WithOrgID(ctx, claims.OrgID)
			ctx = appctx.WithRole(ctx, claims.Role)

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
			role := appctx.GetRole(r.Context())
			if !roleSet[role] {
				http.Error(w, `{"error":"недостаточно прав"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
