package jwt

import (
	"app/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	if duration <= 0 {
		return "", jwt.ErrTokenInvalidClaims
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uid":    user.ID,
			"email":  user.Email,
			"exp":    time.Now().Add(duration).Unix(),
			"app_id": app.ID,
		},
	)

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
