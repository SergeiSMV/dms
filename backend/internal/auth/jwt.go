package auth

import (
	"fmt"
	"time"

	// golang-jwt — библиотека для создания и проверки JSON Web Tokens.
	// JWT позволяет серверу подтвердить личность пользователя без запроса к БД на каждый запрос.
	"github.com/golang-jwt/jwt/v5"
)

// TokenType различает access- и refresh-токены внутри JWT claims.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims — пользовательские данные, которые кодируются внутри JWT.
// RegisteredClaims содержит стандартные поля (exp, iat, sub и др.).
type Claims struct {
	jwt.RegisteredClaims
	UserID string    `json:"user_id"`
	OrgID  string    `json:"org_id"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
}

// JWTManager управляет созданием и проверкой JWT-токенов.
type JWTManager struct {
	secret     []byte        // секретный ключ для подписи (HMAC-SHA256)
	accessTTL  time.Duration // время жизни access-токена
	refreshTTL time.Duration // время жизни refresh-токена
}

// NewJWTManager создаёт менеджер с заданным секретом и временами жизни токенов.
func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccessToken создаёт короткоживущий токен для авторизации API-запросов.
func (m *JWTManager) GenerateAccessToken(userID, orgID, role string) (string, error) {
	return m.generateToken(userID, orgID, role, AccessToken, m.accessTTL)
}

// GenerateRefreshToken создаёт долгоживущий токен для обновления access-токена.
func (m *JWTManager) GenerateRefreshToken(userID, orgID, role string) (string, error) {
	return m.generateToken(userID, orgID, role, RefreshToken, m.refreshTTL)
}

// generateToken — общая логика создания подписанного JWT.
func (m *JWTManager) generateToken(userID, orgID, role string, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Subject — стандартное поле JWT, дублирует UserID для совместимости.
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		Type:   tokenType,
	}

	// jwt.NewWithClaims создаёт неподписанный токен.
	// SignedString подписывает его секретным ключом (HMAC-SHA256).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("подпись токена: %w", err)
	}

	return signed, nil
}

// ParseToken проверяет подпись и срок действия JWT, возвращает claims.
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	// jwt.ParseWithClaims разбирает строку токена, проверяет подпись и expiration.
	// Функция-аргумент возвращает ключ для проверки подписи.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм подписи — HMAC (а не RSA или другой).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("невалидный токен: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("невалидные claims токена")
	}

	return claims, nil
}
