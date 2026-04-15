package jwt

import (
	app "app/internal/auth/app/model"
	user "app/internal/auth/user/model"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	app := app.App{
		ID:     1,
		Secret: "test-secret",
	}

	tests := []struct {
		name     string
		user     user.User
		duration time.Duration
		wantErr  bool
	}{
		{
			name: "Successful token generation",
			user: user.User{
				ID:    1,
				Email: "test@example.com",
			},
			duration: time.Hour,
			wantErr:  false,
		},
		{
			name: "Zero duration",
			user: user.User{
				ID:    3,
				Email: "test3@example.com",
			},
			duration: 0,
			wantErr:  true,
		},
		{
			name: "User with empty email",
			user: user.User{
				ID:    4,
				Email: "",
			},
			duration: time.Hour,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := NewToken(tt.user, app, tt.duration)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, tokenString)

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					t.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(app.Secret), nil
			})
			require.NoError(t, err)

			claims, ok := token.Claims.(jwt.MapClaims)
			require.True(t, ok)

			assert.Equal(t, float64(tt.user.ID), claims["uid"])
			assert.Equal(t, tt.user.Email, claims["email"])
			assert.Equal(t, float64(app.ID), claims["app_id"])

			exp, ok := claims["exp"].(float64)
			require.True(t, ok, "exp claim should be a number")
			expectedExp := time.Now().Add(tt.duration).Unix()
			assert.InDelta(t, expectedExp, int64(exp), 1, "exp claim is not within a second of expected")
		})
	}
}
