// Пакет model содержит доменные сущности приложения.
// Сущности не зависят от БД или HTTP — чистые Go-структуры.
package model

import "time"

// Роли пользователей в системе.
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleViewer  = "viewer"
)

// User — пользователь системы (сотрудник компании-клиента).
type User struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // "-" — никогда не отдаём хеш в JSON
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
