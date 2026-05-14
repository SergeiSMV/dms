// Пакет auth отвечает за аутентификацию: хеширование паролей и работу с JWT-токенами.
package auth

import (
	"fmt"

	// bcrypt — алгоритм хеширования паролей с автоматической солью.
	// Медленный по дизайну: усложняет перебор даже при утечке хешей.
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost — «стоимость» хеширования (2^cost итераций).
// 12 — хороший баланс безопасности и скорости (~250мс на современном CPU).
const bcryptCost = 12

// HashPassword создаёт bcrypt-хеш из открытого пароля.
// Соль генерируется автоматически и встраивается в результат.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("хеширование пароля: %w", err)
	}
	return string(hash), nil
}

// CheckPassword сравнивает открытый пароль с его bcrypt-хешем.
// Возвращает nil, если пароль верный, или ошибку, если нет.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
