// Пакет service содержит бизнес-логику приложения.
// Сервисы координируют работу репозиториев, auth и других слоёв.
package service

import (
	"context"
	"errors"
	"fmt"

	"dms-backend/internal/appctx"
	"dms-backend/internal/auth"
	"dms-backend/internal/model"
	"dms-backend/internal/repository"
)

// Ошибки аутентификации — возвращаются хендлерам для выбора HTTP-статуса.
var (
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	ErrUserInactive       = errors.New("пользователь деактивирован")
	ErrInvalidToken       = errors.New("невалидный токен")
	ErrEmailExists        = errors.New("пользователь с таким email уже существует")
)

// TokenPair — пара access + refresh токенов, возвращаемая при логине.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthService — сервис аутентификации: логин, refresh, смена пароля, создание пользователей.
type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService создаёт сервис с зависимостями.
func NewAuthService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Login проверяет email/пароль и возвращает пару токенов.
// org_id должен быть в контексте (через appctx.WithOrgID).
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Сравниваем введённый пароль с bcrypt-хешем из БД.
	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

// RefreshTokens принимает refresh-токен и выдаёт новую пару.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Принимаем только refresh-токены, а не access.
	if claims.Type != auth.RefreshToken {
		return nil, ErrInvalidToken
	}

	// Кладём org_id из refresh-токена в контекст для репозитория.
	ctx = appctx.WithOrgID(ctx, claims.OrgID)

	// Проверяем, что пользователь всё ещё активен.
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return s.generateTokenPair(user)
}

// ChangePassword меняет пароль текущего пользователя.
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	// Проверяем старый пароль.
	if err := auth.CheckPassword(oldPassword, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}

	// Хешируем новый пароль и сохраняем.
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, userID, hash)
}

// GetProfile возвращает профиль пользователя по ID.
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// CreateUser создаёт нового пользователя (только admin).
// org_id берётся из контекста.
func (s *AuthService) CreateUser(ctx context.Context, email, password, role string) (*model.User, error) {
	// Проверяем, не занят ли email.
	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("создание пользователя: %w", err)
	}

	return user, nil
}

// ListUsers возвращает всех пользователей организации из контекста.
func (s *AuthService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}

// generateTokenPair создаёт пару access + refresh токенов для пользователя.
func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.OrgID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.OrgID, user.Role)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
