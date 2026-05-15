// Пакет appctx — доступ к данным текущего пользователя и организации через context.
// Middleware записывает значения при обработке JWT, репозитории и сервисы читают.
package appctx

import "context"

// contextKey — тип для ключей контекста, чтобы избежать коллизий со строками.
type contextKey string

const (
	userIDKey contextKey = "user_id"
	orgIDKey  contextKey = "org_id"
	roleKey   contextKey = "role"
)

// WithUserID добавляет ID пользователя в контекст.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// WithOrgID добавляет ID организации в контекст.
func WithOrgID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, orgIDKey, id)
}

// WithRole добавляет роль пользователя в контекст.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// GetUserID извлекает ID пользователя из контекста.
func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

// GetOrgID извлекает ID организации из контекста.
func GetOrgID(ctx context.Context) string {
	v, _ := ctx.Value(orgIDKey).(string)
	return v
}

// GetRole извлекает роль пользователя из контекста.
func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(roleKey).(string)
	return v
}
