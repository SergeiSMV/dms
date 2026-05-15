// Пакет server создаёт и настраивает HTTP-сервер приложения.
package server

import (
	"net/http"
	"time"

	// chi — легковесный роутер для Go, совместимый со стандартным net/http.
	"github.com/go-chi/chi/v5"
	// Набор готовых middleware: логирование запросов, recovery от panic и др.
	chimw "github.com/go-chi/chi/v5/middleware"

	"dms-backend/internal/auth"
	appmw "dms-backend/internal/middleware"
	"dms-backend/internal/model"
	"dms-backend/internal/service"
)

// Deps — зависимости, которые нужны серверу для создания маршрутов.
type Deps struct {
	AuthService  *service.AuthService
	OrgService   *service.OrgService
	JWTManager   *auth.JWTManager
	DefaultOrgID string // ID организации по умолчанию (on-premise — всегда одна)
}

// New создаёт и возвращает настроенный HTTP-сервер.
// addr — адрес вида ":8081", на котором сервер будет слушать.
func New(addr string, deps Deps) *http.Server {
	r := chi.NewRouter()

	// --- Глобальные middleware ---

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))

	// --- Публичные маршруты (без авторизации) ---

	r.Get("/health", handleHealth())

	// Аутентификация: логин и обновление токена не требуют JWT.
	r.Post("/auth/login", handleLogin(deps.AuthService, deps.DefaultOrgID))
	r.Post("/auth/refresh", handleRefresh(deps.AuthService))

	// --- Защищённые маршруты (требуют access-токен) ---

	r.Group(func(r chi.Router) {
		// Auth middleware проверяет JWT и кладёт user_id/org_id/role в контекст.
		r.Use(appmw.Auth(deps.JWTManager))

		r.Get("/auth/profile", handleProfile(deps.AuthService))
		r.Post("/auth/change-password", handleChangePassword(deps.AuthService))

		// Организация: доступна всем авторизованным, обновление — только admin.
		r.Get("/organization", handleGetOrganization(deps.OrgService))

		// Admin-маршруты: только для роли admin.
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireRole(model.RoleAdmin))

			r.Patch("/organization", handleUpdateOrganization(deps.OrgService))
			r.Post("/admin/users", handleCreateUser(deps.AuthService))
			r.Get("/admin/users", handleListUsers(deps.AuthService))
		})
	})

	return &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
