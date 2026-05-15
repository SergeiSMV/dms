package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"dms-backend/internal/appctx"
	"dms-backend/internal/model"
	"dms-backend/internal/service"
)

// handleHealth возвращает хендлер для GET /health.
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// --- Auth хендлеры ---

// loginRequest — тело запроса POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleLogin — POST /auth/login. Проверяет email/пароль, возвращает пару JWT-токенов.
func handleLogin(authService *service.AuthService, defaultOrgID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
			return
		}

		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email и password обязательны"})
			return
		}

		// На on-premise всегда одна организация — кладём defaultOrgID в контекст,
		// чтобы репозиторий мог автоматически фильтровать по org_id.
		ctx := appctx.WithOrgID(r.Context(), defaultOrgID)

		tokens, err := authService.Login(ctx, req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidCredentials):
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "неверный email или пароль"})
			case errors.Is(err, service.ErrUserInactive):
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "пользователь деактивирован"})
			default:
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка"})
			}
			return
		}

		writeJSON(w, http.StatusOK, tokens)
	}
}

// refreshRequest — тело запроса POST /auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// handleRefresh — POST /auth/refresh. Принимает refresh-токен, выдаёт новую пару.
// org_id извлекается из refresh-токена и кладётся в контекст для репозитория.
func handleRefresh(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
			return
		}

		if req.RefreshToken == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token обязателен"})
			return
		}

		tokens, err := authService.RefreshTokens(r.Context(), req.RefreshToken)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "невалидный refresh-токен"})
			return
		}

		writeJSON(w, http.StatusOK, tokens)
	}
}

// handleProfile — GET /auth/profile. Профиль текущего пользователя (из JWT).
func handleProfile(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := appctx.GetUserID(r.Context())

		user, err := authService.GetProfile(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "пользователь не найден"})
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

// changePasswordRequest — тело запроса POST /auth/change-password.
type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// handleChangePassword — POST /auth/change-password.
func handleChangePassword(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req changePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
			return
		}

		if req.OldPassword == "" || req.NewPassword == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "old_password и new_password обязательны"})
			return
		}

		userID := appctx.GetUserID(r.Context())
		err := authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "неверный текущий пароль"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "пароль изменён"})
	}
}

// --- Admin хендлеры ---

// createUserRequest — тело запроса POST /admin/users.
type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// handleCreateUser — POST /admin/users. Создание пользователя (только admin).
// org_id берётся из контекста (положен JWT middleware).
func handleCreateUser(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
			return
		}

		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email и password обязательны"})
			return
		}

		// Валидация роли.
		role := req.Role
		if role == "" {
			role = model.RoleViewer
		}
		if role != model.RoleAdmin && role != model.RoleManager && role != model.RoleViewer {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "допустимые роли: admin, manager, viewer"})
			return
		}

		user, err := authService.CreateUser(r.Context(), req.Email, req.Password, role)
		if err != nil {
			if errors.Is(err, service.ErrEmailExists) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "пользователь с таким email уже существует"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка"})
			return
		}

		writeJSON(w, http.StatusCreated, user)
	}
}

// handleListUsers — GET /admin/users. Список пользователей организации.
// org_id берётся из контекста (положен JWT middleware).
func handleListUsers(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := authService.ListUsers(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка"})
			return
		}

		writeJSON(w, http.StatusOK, users)
	}
}

// --- Organization хендлеры ---

// handleGetOrganization — GET /organization. Текущая организация пользователя.
func handleGetOrganization(orgService *service.OrgService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org, err := orgService.Get(r.Context())
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "организация не найдена"})
			return
		}

		writeJSON(w, http.StatusOK, org)
	}
}

// updateOrgRequest — тело запроса PATCH /organization.
// Указатели позволяют отличить «не передано» от «пустая строка».
type updateOrgRequest struct {
	Name     *string          `json:"name"`
	INN      *string          `json:"inn"`
	Settings *json.RawMessage `json:"settings"`
}

// handleUpdateOrganization — PATCH /organization. Обновление данных организации (только admin).
func handleUpdateOrganization(orgService *service.OrgService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req updateOrgRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "невалидный JSON"})
			return
		}

		if req.Name == nil && req.INN == nil && req.Settings == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "нет полей для обновления"})
			return
		}

		org, err := orgService.Update(r.Context(), req.Name, req.INN, req.Settings)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "внутренняя ошибка"})
			return
		}

		writeJSON(w, http.StatusOK, org)
	}
}

// --- Утилиты ---

// writeJSON отправляет JSON-ответ с заданным статус-кодом.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
